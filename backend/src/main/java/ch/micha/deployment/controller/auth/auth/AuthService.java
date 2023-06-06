package ch.micha.deployment.controller.auth.auth;

import ch.micha.deployment.controller.auth.EncodingUtil;
import ch.micha.deployment.controller.auth.entity.credentials.Credential;
import ch.micha.deployment.controller.auth.entity.page.Page;
import ch.micha.deployment.controller.auth.entity.user.User;
import io.helidon.common.http.Http.Status;
import io.helidon.config.Config;
import io.helidon.dbclient.DbClient;
import io.helidon.dbclient.DbRow;
import io.helidon.webserver.Handler;
import io.helidon.webserver.RequestHeaders;
import io.helidon.webserver.Routing.Rules;
import io.helidon.webserver.ServerRequest;
import io.helidon.webserver.ServerResponse;
import io.helidon.webserver.Service;
import io.jsonwebtoken.Claims;
import io.jsonwebtoken.Jwts;
import io.jsonwebtoken.SignatureAlgorithm;
import java.security.Key;
import java.time.Instant;
import java.time.temporal.ChronoUnit;
import java.util.Base64;
import java.util.Date;
import java.util.Optional;
import java.util.logging.Level;
import java.util.logging.Logger;
import javax.crypto.spec.SecretKeySpec;

public class AuthService implements Service {
    private static final Logger LOGGER = Logger.getLogger(AuthService.class.getSimpleName());
    private static final SignatureAlgorithm SIGNATURE_ALGORITHM = SignatureAlgorithm.HS512;
    private static final String AUTH_REQUEST_PAGE_QUERY = "page";
    private static final String BEARER_COOKIE = "Bearer";

    private final DbClient db;
    private final Key key;
    private final long tokenExpireHours;

    public AuthService(DbClient db, Config config) {
        this.db = db;

        String keyConfig = config.get("key").asString().get();
        key = new SecretKeySpec(Base64.getDecoder().decode(keyConfig), SIGNATURE_ALGORITHM.getJcaName());

        tokenExpireHours = config.get("tokenExpireHours").asLong().get();

        createDefaultAdmin(config.get("default"));
    }

    @Override
    public void update(Rules rules) {
        rules
            .post("/login", Handler.create(Credential.class, this::login))
            .post("/auth", this::validateTokenCookie);
    }

    public void login(ServerRequest request, ServerResponse response, Credential credential) {
        LOGGER.log(Level.INFO, "trying login for {0}", credential.mail());

        User user = selectUserByMail(credential.mail());
        if(user == null) {
            LOGGER.log(Level.WARNING, "login denied for {0}", new Object[]{ credential.mail() });
            response.status(Status.FORBIDDEN_403);
            response.send("invalid credentials");
            return;
        }

        String hashedPassword = EncodingUtil.hashString(credential.password(), user.salt());
        if(hashedPassword.equals(user.password())) {
            LOGGER.log(Level.INFO, "login granted for {0}", credential.mail());

            String token = createJwt(new SecurityToken(
                request.remoteAddress(),
                Date.from(Instant.now()),
                user.mail(),
                user.admin(),
                user.viewPrivate(),
                Date.from(Instant.now().plus(tokenExpireHours, ChronoUnit.HOURS))
            ));
            response.status(Status.NO_CONTENT_204);
            response.addHeader("Set-Cookie", String.format("%s=%s; Path=/; HttpOnly=true;", BEARER_COOKIE, token));
            response.send();
        } else {
            LOGGER.log(Level.WARNING, "login denied for {0}", new Object[]{ credential.mail() });
            response.status(Status.FORBIDDEN_403);
            response.send("invalid credentials");
        }
    }

    public void validateTokenCookie(ServerRequest request, ServerResponse response) {
        Optional<String> pageIdQuery = request.queryParams().first(AUTH_REQUEST_PAGE_QUERY);

        if(pageIdQuery.isEmpty()) {
            LOGGER.log(Level.INFO, "missing param at auth request, from ip: {0}",
                new Object[]{ request.remoteAddress() });
            response.status(Status.BAD_REQUEST_400);
            response.send("missing parameter");
            return;
        }

        LOGGER.log(Level.INFO, "validating token for request to page {0}", new Object[]{ pageIdQuery.get() });
        int pageId = Integer.parseInt(pageIdQuery.get());
        DbRow pageRow = db.execute(exec -> exec
                .createNamedQuery("select-page")
                .addParam("id", pageId)
                .execute()
            ).first()
            .await();

        if(pageRow == null) {
            LOGGER.log(Level.INFO, "access to unknown page denied, from: {0}",
                new Object[]{ request.remoteAddress() });
            response.status(Status.FORBIDDEN_403);
            response.send("not allowed");
            return;
        }

        Page page = pageRow.as(Page.class);
        if(!page.privateAccess()) {
            LOGGER.log(Level.INFO, "access to public page granted: {0}",
                new Object[]{ page.url() });
            response.status(Status.OK_200);
            response.send("enjoy!");
            return;
        }

        SecurityToken token = extractTokenCookie(request.headers(), response);
        validateSecurityToken(request, token, response);
        if(token.isPrivateAccess()) {
            LOGGER.log(Level.INFO, "access to private page {0} granted for {1}",
                new Object[]{ page.url(), token.getUserMail() });
            response.status(Status.OK_200);
            response.send("enjoy!");
        } else {
            LOGGER.log(Level.WARNING, "access to private page {0} refused for {1}",
                new Object[]{ page.url(), token.getUserMail() });
            response.status(Status.FORBIDDEN_403);
            response.send("not allowed");
        }
    }

    public String createJwt(SecurityToken securityToken) {
        return Jwts.builder()
            .setIssuer(securityToken.getIssuer())
            .setIssuedAt(securityToken.getIssuedAt())
            .setExpiration(securityToken.getExpiresAt())
            .claim(SecurityToken.CLAIM_USER_MAIL, securityToken.getUserMail())
            .claim(SecurityToken.CLAIM_ADMIN, securityToken.isAdmin())
            .claim(SecurityToken.CLAIM_PRIVATE_ACCESS, securityToken.isPrivateAccess())
            .signWith(SIGNATURE_ALGORITHM, key)
            .compact();
    }

    public SecurityToken parseJwt(String token) {
        SecurityToken securityToken = new SecurityToken();
        Claims claims = Jwts.parser().setSigningKey(key).parseClaimsJws(token).getBody();

        securityToken.setIssuer(claims.getIssuer());
        securityToken.setIssuedAt(claims.getIssuedAt());
        securityToken.setUserMail(claims.get(SecurityToken.CLAIM_USER_MAIL, String.class));
        securityToken.setAdmin(claims.get(SecurityToken.CLAIM_ADMIN, Boolean.class));
        securityToken.setPrivateAccess(claims.get(SecurityToken.CLAIM_PRIVATE_ACCESS, Boolean.class));
        securityToken.setExpiresAt(claims.getExpiration());

        return securityToken;
    }

    public SecurityToken extractTokenCookie(RequestHeaders headers, ServerResponse response) {
        Optional<String> tokenCookie = headers.cookies().first(BEARER_COOKIE);

        if(tokenCookie.isEmpty()) {
            LOGGER.log(Level.INFO, "got admin request with no auth cookie");
            response.status(Status.UNAUTHORIZED_401);
            response.send("Unauthorized request");
            return null;
        }

        return parseJwt(tokenCookie.get());
    }

    public void validateSecurityToken(ServerRequest request, SecurityToken token, ServerResponse response) {
        // handle client change (token rubbery dude)
        if(!token.getIssuer().equals(request.remoteAddress())) {
            LOGGER.log(Level.WARNING, "invalid issuer in request: {0} changed to {1}, associated user: {2}",
                new Object[]{ token.getIssuer(), request.remoteAddress(), token.getUserMail() });
            response.status(Status.UNAUTHORIZED_401);
            response.send("Unauthorized request");
            // handle token expired
        } else if(token.getExpiresAt().before(Date.from(Instant.now()))) {
            LOGGER.log(Level.WARNING, "token for {0} expired",
                new Object[]{ token.getUserMail() });
            response.status(Status.UNAUTHORIZED_401);
            response.send("token expired");
        }
    }

    private void createDefaultAdmin(Config defaultConfig) {
        String defaultMail = defaultConfig.get("mail").as(String.class).get();
        String defaultPassword = defaultConfig.get("password").as(String.class).get();
        LOGGER.log(Level.INFO, "checking if default user with mail {0} exists", defaultMail);

        User existingDefaultUser = selectUserByMail(defaultMail);
        if(existingDefaultUser == null) {
            LOGGER.log(Level.INFO, "default user not found -> creating one");

            String salt = EncodingUtil.generateSalt();
            String hashedPassword = EncodingUtil.hashString(defaultPassword, salt);

            db.execute(exec -> exec
                    .createNamedInsert("insert-user")
                    .addParam("mail", defaultMail)
                    .addParam("password", hashedPassword)
                    .addParam("salt", salt)
                    .addParam("admin", true)
                    .addParam("view_private", true)
                    .execute())
                .thenAccept(count -> LOGGER.log(Level.INFO, "created default user"))
                .exceptionally(t -> {
                    LOGGER.log(Level.SEVERE, "failed to create default user", t);
                    return null;
                });
        }
    }

    private User selectUserByMail(String mail) {
        DbRow row = db.execute(exec -> exec
                .createNamedQuery("select-user-mail")
                .addParam("mail", mail)
                .execute()
            ).first()
            .await();
        if(row == null)
            return null;
        return row.as(User.class);
    }
}
