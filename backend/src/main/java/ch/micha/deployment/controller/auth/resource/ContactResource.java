package ch.micha.deployment.controller.auth.resource;

import ch.micha.deployment.controller.auth.dto.ContactDto;
import ch.micha.deployment.controller.auth.logging.RequestLogHandler;
import ch.micha.deployment.controller.auth.mail.SendMailDto;
import ch.micha.deployment.controller.auth.mail.SendMailDto.Type;
import io.helidon.common.http.Http.Status;
import io.helidon.webserver.Handler;
import io.helidon.webserver.Routing.Rules;
import io.helidon.webserver.ServerRequest;
import io.helidon.webserver.ServerResponse;
import io.helidon.webserver.Service;
import java.util.concurrent.BlockingQueue;
import java.util.logging.Level;
import java.util.logging.Logger;

public class ContactResource implements Service {
    private static final Logger LOGGER = Logger.getLogger(ContactResource.class.getSimpleName());

    private final BlockingQueue<SendMailDto> sendMailQueue;
    private final String adminMail;

    public ContactResource(BlockingQueue<SendMailDto> sendMailQueue, String adminMail) {
        this.sendMailQueue = sendMailQueue;
        this.adminMail = adminMail;
    }

    @Override
    public void update(Rules rules) {
        rules.post("/", Handler.create(ContactDto.class, this::postContact));
    }

    private void postContact(ServerRequest request, ServerResponse response, ContactDto contact) {
        final String requestId = RequestLogHandler.parseRequestId(request);
        LOGGER.log(Level.FINE, "{0} sending contact request (length:{1}) from {2}", new Object[]{ requestId, contact.message().length(), contact.mail() });
        sendMailQueue.add(new SendMailDto(Type.CONTACT_REQUEST, contact, adminMail));
        response.status(Status.CREATED_201).send();
    }
}
