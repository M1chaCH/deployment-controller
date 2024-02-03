package ch.micha.deployment.controller.auth.mail;

import ch.micha.deployment.controller.auth.auth.SecurityToken;
import ch.micha.deployment.controller.auth.location.LocationResolver;
import com.maxmind.geoip2.model.CityResponse;
import io.helidon.config.Config;
import jakarta.mail.Authenticator;
import jakarta.mail.Message;
import jakarta.mail.Message.RecipientType;
import jakarta.mail.MessagingException;
import jakarta.mail.PasswordAuthentication;
import jakarta.mail.Session;
import jakarta.mail.Transport;
import jakarta.mail.internet.AddressException;
import jakarta.mail.internet.InternetAddress;
import jakarta.mail.internet.MimeBodyPart;
import jakarta.mail.internet.MimeMessage;
import jakarta.mail.internet.MimeMultipart;
import java.util.Optional;
import java.util.Properties;
import java.util.concurrent.BlockingQueue;
import java.util.logging.Level;
import java.util.logging.Logger;
import org.eclipse.angus.mail.util.MailConnectException;

public class SendMailProcessor implements Runnable{
    private static final Logger LOGGER = Logger.getLogger(SendMailProcessor.class.getSimpleName());
    private final BlockingQueue<SendMailDto> sendMailQueue;
    private final Session session;
    private final String mailFrom;
    private final LocationResolver locationResolver;

    public SendMailProcessor(BlockingQueue<SendMailDto> sendMailQueue, Config appConfig) {
        this.sendMailQueue = sendMailQueue;
        Config mailConfig = appConfig.get("security.mail");
        this.mailFrom = mailConfig.get("from").asString().get();
        LOGGER.log(Level.INFO, "initializing mail processor for {0}", new Object[]{ mailFrom });

        String smtpServer = mailConfig.get("smtp").get("server").asString().get();
        String smtpPort = mailConfig.get("smtp").get("port").asString().get();
        Properties smtpProps = new Properties();
        smtpProps.put("mail.smtp.auth", true);
        smtpProps.put("mail.smtp.starttls.enable", "true");
        smtpProps.put("mail.smtp.host", smtpServer);
        smtpProps.put("mail.smtp.port", smtpPort);
        smtpProps.put("mail.smtp.ssl.trust", smtpServer);

        session = Session.getInstance(smtpProps, new Authenticator() {
            @Override
            protected PasswordAuthentication getPasswordAuthentication() {
                return new PasswordAuthentication(
                    mailConfig.get("smtp").get("user").asString().get(),
                    mailConfig.get("smtp").get("password").asString().get()
                );
            }
        });

        this.locationResolver = LocationResolver.getInstance(appConfig.get("location"));
    }

    @SuppressWarnings({"java:S2189", "InfiniteLoopStatement"}) // it makes sense for this loop to be infinite
    @Override
    public void run() {
        Thread.currentThread().setName("mail-sender");
        LOGGER.log(Level.INFO, "mail sender thread created and started: {0}", new Object[]{ Thread.currentThread().getName() });

        try {
            while (true)
                listenForMail();

        } catch (InterruptedException e){
            LOGGER.log(Level.WARNING, "{0} interrupted -> re-interrupting", new Object[]{ Thread.currentThread().getName() });
            Thread.currentThread().interrupt();
        }
    }

    private void listenForMail() throws InterruptedException {
        try {
            LOGGER.log(Level.FINE, "waiting for mail");
            SendMailDto toSend = sendMailQueue.take();
            LOGGER.log(Level.FINE, "got mail: {0}", new Object[]{ toSend.getMailType().name() });

            String subject = switch (toSend.getMailType()) {
                case LOGIN_GRANT -> "Deployment Controller: login granted";
                case PAGE_INVITATION -> "michus-net Invitation";
                case USER_ACTIVATED -> "Deployment Controller: user activated";
            };

            Message message = new MimeMessage(session);
            message.setFrom(new InternetAddress(mailFrom));
            message.setRecipient(RecipientType.TO, new InternetAddress(toSend.getRecipient()));
            message.setSubject(subject);

            MimeBodyPart body = new MimeBodyPart();
            body.setContent(createHtmlBody(toSend), "text/html; charset=utf-8");

            message.setContent(new MimeMultipart(body));
            Transport.send(message);
            LOGGER.log(Level.FINE, "successfully sent mail", new Object[]{ });
        } catch (AddressException | MailConnectException e) {
            LOGGER.log(Level.SEVERE, "invalid address exception, stopping mail process!", e);
            // interrupt because these are configure values, if they are wrong, it won't get better over time (:
            Thread.currentThread().interrupt();
        } catch (MessagingException e) {
            LOGGER.log(Level.WARNING, "failed to send mail", e);
        }
    }

    private String createHtmlBody(SendMailDto toSend) {
        String defaultBody = "unknown data for mail";
        return switch (toSend.getMailType()) {
            case LOGIN_GRANT -> {
                if(toSend.getData() instanceof SecurityToken token)
                   yield buildLoginMessage(token);
                yield defaultBody;
            }
            case PAGE_INVITATION -> {
                if(toSend.getData() instanceof PageInvitationMailDto pageInvitation)
                    yield buildPageInvitationMessage(pageInvitation);
                yield defaultBody;
            }
            case USER_ACTIVATED -> {
                if(toSend.getData() instanceof  UserActivatedMailDto userActivated)
                    yield buildUserActivatedMessage(userActivated);
                yield defaultBody;
            }
        };
    }

    private String buildUserActivatedMessage(UserActivatedMailDto userActivated) {
        return """
               <h3>Hi</h3>
               <p>Just letting you know, a user has just activated their account.</p>
               <p>User: %s</p>
               <p>Time: %s</p>
               <br />
               <p>Regards - michu de dev 🙋‍♂️</p>
               """.formatted(userActivated.getUserMail(), userActivated.getTime());
    }

    private String buildPageInvitationMessage(PageInvitationMailDto pageInvitation) {
        StringBuilder pages = new StringBuilder("<ul>");
        for (String page : pageInvitation.getPages()) {
            pages.append("<li>");
            pages.append(page);
            pages.append("</li>");
        }
        pages.append("</ul>");

        return """
               <h3>Hi %s</h3>
               <p>You have been invited to access pages on michus-net.</p>
               <p>Please head to the link <a href="%s">here</a> and complete the onboarding step.</p>
               <br />
               <p>You were invited visit the following pages:</p>
               %s
               <br />
               <p>Best Regards</p>
               <p>Micha</p>
               """.formatted(pageInvitation.getUserMail().split("@")[0], pageInvitation.getUrl(), pages);
    }

    private String buildLoginMessage(SecurityToken token) {
        String message = """
            <h3>Hi</h3>
            <p>Deployment Controller has granted a new login.</p>
            <ul>
                <li>Source: %s, %s, %s</li>
                <li>User: %s</li>
                <li>Admin: %s</li>
                <li>Private access to: %s</li>
                <li>Time: %s</li>
            </ul>
            <p>If you did not send this message, please go ahead and shut down our environment.</p>
            <p>Regards - michu de dev 🙋‍♂️</p>
            """;

        Optional<CityResponse> location = locationResolver.resolveLocation(token.getIssuer());
        String country = "unknown";
        String city = "unknown";
        if(location.isPresent()) {
            country = location.get().getCountry().getName();
            city = location.get().getCity().getName();
        }

        return String.format(message,
                             token.getIssuer(), country, city,
                             token.getUserMail(),
                             token.isAdmin(),
                             token.getPrivatePagesAccess().replace(SecurityToken.CLAIM_PRIVATE_ACCESS_DELIMITER, ", "),
                             token.getIssuedAt().toString()
        );
    }
}
