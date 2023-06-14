package ch.micha.deployment.controller.auth.auth;

import ch.micha.deployment.controller.auth.EncodingUtil;
import ch.micha.deployment.controller.auth.entity.credentials.Credential;
import ch.micha.deployment.controller.auth.entity.page.Page;
import ch.micha.deployment.controller.auth.entity.user.User;
import ch.micha.deployment.controller.auth.error.BadRequestException;
import ch.micha.deployment.controller.auth.error.ForbiddenException;
import ch.micha.deployment.controller.auth.error.UnauthorizedException;
import ch.micha.deployment.controller.auth.logging.RequestLogHandler;
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
import jakarta.json.Json;
import jakarta.json.JsonObjectBuilder;
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
    private static final String AUTH_REQUEST_PAGE_PARAM = "page";
    private static final String BEARER_COOKIE = "Bearer";

    private final DbClient db;
    private final Config securityConfig;
    private final Key key;
    private final long tokenExpireHours;

    public AuthService(DbClient db, Config config) {
        this.db = db;
        this.securityConfig = config;

        String keyConfig = config.get("key").asString().get();
        key = new SecretKeySpec(Base64.getDecoder().decode(keyConfig), SIGNATURE_ALGORITHM.getJcaName());

        tokenExpireHours = config.get("tokenExpireHours").asLong().get();

        createDefaults(config.get("default"));
    }

    @Override
    public void update(Rules rules) {
        rules
            .post("/login", Handler.create(Credential.class, this::login))
            .get("/auth/{" + AUTH_REQUEST_PAGE_PARAM + "}", this::validateTokenCookie)
            .get("/auth", this::isLoggedIn);
    }

    public void login(ServerRequest request, ServerResponse response, Credential credential) {
        final String requestId = RequestLogHandler.parseRequestId(request);
        LOGGER.log(Level.FINE, "{0} trying login for {1}", new Object[]{ requestId, credential.mail() });

        User user = selectUserByMail(credential.mail());
        if(user == null)
            throw new ForbiddenException(String.format("login denied for %s", credential.mail()), "invalid credentials");

        String hashedPassword = EncodingUtil.hashString(credential.password(), user.salt());
        if(hashedPassword.equals(user.password())) {
            LOGGER.log(Level.FINE, "{0} login granted for {1}", new Object[]{requestId, credential.mail()});

            String token = createJwt(new SecurityToken(
                request.remoteAddress(),
                Date.from(Instant.now()),
                user.mail(),
                user.admin(),
                user.viewPrivate(),
                Date.from(Instant.now().plus(tokenExpireHours, ChronoUnit.HOURS))
            ));

            String domain = securityConfig.get("domain").asString().get();

            response.status(Status.NO_CONTENT_204);
            response.addHeader("Set-Cookie", String.format("%s=%s; Path=/; HttpOnly=true; "
                + "SameSite=Strict; Secure=true; Domain=%s;",
                BEARER_COOKIE, token, domain));
            response.send();
        } else
            throw new ForbiddenException(String.format("login denied for %s", credential.mail()), "invalid credentials");
    }

    public void validateTokenCookie(ServerRequest request, ServerResponse response) {
        final String requestId = RequestLogHandler.parseRequestId(request);
        String pageIdParam = request.path().param(AUTH_REQUEST_PAGE_PARAM);

        if(pageIdParam == null || pageIdParam.isEmpty() || pageIdParam.isBlank())
            throw new BadRequestException(String.format("missing page param at auth request, from ip: %s",
                request.remoteAddress()), "missing parameter");

        LOGGER.log(Level.FINE, "{0} validating token for request to page {1}", new Object[]{ requestId, pageIdParam });

        Page page = selectPageById(pageIdParam);
        if(page == null)
            throw new ForbiddenException(String.format("access to unknown page denied, from: %s",
                request.remoteAddress()), "not allowed");

        if(!page.privateAccess()) {
            LOGGER.log(Level.FINE, "{0} access to public page granted: {1}",
                new Object[]{ requestId, page.url() });
            response.status(Status.OK_200);
            response.send("enjoy!");
            return;
        }

        SecurityToken token = extractTokenCookie(request.headers());
        validateSecurityToken(request, token);
        if(token.isPrivateAccess()) {
            LOGGER.log(Level.FINE, "{0} access to private page {1} granted for {2}",
                new Object[]{ requestId, page.url(), token.getUserMail() });
            response.status(Status.OK_200);
            response.send("enjoy!");
        } else
            throw new ForbiddenException(String.format("access to private page %s refused for %s",
                page.url(), token.getUserMail()), "not allowed");
    }

    public void isLoggedIn(ServerRequest request, ServerResponse response) {
        SecurityToken token = extractTokenCookie(request.headers());
        User user = selectUserByMail(token.getUserMail());

        if(user == null) {
            response.status(Status.NO_CONTENT_204);
            response.send();
            return;
        }

        JsonObjectBuilder jsonUserBuilder = Json.createObjectBuilder();
        jsonUserBuilder.add("mail", user.mail());
        jsonUserBuilder.add("admin", user.admin());
        jsonUserBuilder.add("viewPrivate", user.viewPrivate());

        response.send(jsonUserBuilder.build());
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

    public SecurityToken extractTokenCookie(RequestHeaders headers) {
        Optional<String> tokenCookie = headers.cookies().first(BEARER_COOKIE);

        if(tokenCookie.isEmpty())
            throw new UnauthorizedException("got admin request with no auth cookie", "unauthorized request");

        return parseJwt(tokenCookie.get());
    }

    public void validateSecurityToken(ServerRequest request, SecurityToken token) {
        if(token == null)
            throw new UnauthorizedException(String.format("got request from %s, with no token provided",
                request.remoteAddress()), "unauthorized");
        // handle client change (token rubbery dude)
        if(!token.getIssuer().equals(request.remoteAddress())) {
            throw new UnauthorizedException(String.format("invalid issuer in request: %s changed to %s, associated user: %s",
                token.getIssuer(), request.remoteAddress(), token.getUserMail()), "unauthorized request");
            // handle token expired
        } else if(token.getExpiresAt().before(Date.from(Instant.now()))) {
            throw new UnauthorizedException(String.format("token for %s expired",
                token.getUserMail()), "token expired");
        }
    }

    private void createDefaults(Config defaultConfig) {
        String defaultMail = defaultConfig.get("mail").as(String.class).get();
        String defaultPassword = defaultConfig.get("password").as(String.class).get();

        createDefaultAdmin(defaultMail, defaultPassword);
    }

    private void createDefaultAdmin(String defaultMail, String defaultPassword) {
        LOGGER.log(Level.FINE, "checking if default user with mail {0} exists", defaultMail);

        User existingDefaultUser = selectUserByMail(defaultMail);
        if(existingDefaultUser == null) {
            LOGGER.log(Level.FINE, "default user not found -> creating one");

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
                .thenAccept(count -> LOGGER.log(Level.FINE, "created default user"))
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

    private Page selectPageById(String id) {
        DbRow row = db.execute(exec -> exec
                .createNamedQuery("select-page")
                .addParam("id", id)
                .execute()
            ).first()
            .await();
        if(row == null)
            return null;
        return row.as(Page.class);
    }
}
