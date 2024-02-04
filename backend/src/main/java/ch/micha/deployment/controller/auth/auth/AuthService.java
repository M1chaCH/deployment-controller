package ch.micha.deployment.controller.auth.auth;

import ch.micha.deployment.controller.auth.EncodingUtil;
import ch.micha.deployment.controller.auth.db.CachedPageDb;
import ch.micha.deployment.controller.auth.db.CachedUserDb;
import ch.micha.deployment.controller.auth.db.PageEntity;
import ch.micha.deployment.controller.auth.db.UserEntity;
import ch.micha.deployment.controller.auth.db.UserPageEntity;
import ch.micha.deployment.controller.auth.dto.ChangeCredentialDto;
import ch.micha.deployment.controller.auth.dto.CredentialDto;
import ch.micha.deployment.controller.auth.dto.UserReadDto;
import ch.micha.deployment.controller.auth.error.BadRequestException;
import ch.micha.deployment.controller.auth.error.ForbiddenException;
import ch.micha.deployment.controller.auth.error.InactiveUserException;
import ch.micha.deployment.controller.auth.error.UnauthorizedException;
import ch.micha.deployment.controller.auth.logging.RequestLogHandler;
import ch.micha.deployment.controller.auth.mail.SendMailDto;
import ch.micha.deployment.controller.auth.mail.SendMailDto.Type;
import ch.micha.deployment.controller.auth.mail.UserActivatedMailDto;
import io.helidon.common.http.Http.Status;
import io.helidon.config.Config;
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
import java.sql.SQLException;
import java.time.Instant;
import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.time.temporal.ChronoUnit;
import java.util.Base64;
import java.util.Date;
import java.util.Optional;
import java.util.UUID;
import java.util.concurrent.BlockingQueue;
import java.util.logging.Level;
import java.util.logging.Logger;
import java.util.stream.Collectors;
import javax.crypto.spec.SecretKeySpec;

public class AuthService implements Service {
    private static final Logger LOGGER = Logger.getLogger(AuthService.class.getSimpleName());
    private static final SignatureAlgorithm SIGNATURE_ALGORITHM = SignatureAlgorithm.HS512;
    private static final String AUTH_REQUEST_PAGE_PARAM = "page";
    private static final String BEARER_COOKIE = "Bearer";

    private final CachedUserDb userDb;
    private final CachedPageDb pageDb;
    private final Config securityConfig;
    private final Key key;
    private final long tokenExpireHours;
    private final BlockingQueue<SendMailDto> sendMailQueue;
    private final String adminMail;

    public static String loadRemoteAddress(ServerRequest request) {
        Optional<String> xRealIp = request.headers().first("X-Real-IP");
        if(xRealIp.isPresent() && !xRealIp.get().isBlank())
            return xRealIp.get();

        Optional<String> xForwardFor = request.headers().first("X-Forwarded-For");
        if(xForwardFor.isPresent() && !xForwardFor.get().isBlank())
            return xForwardFor.get();

        return request.remoteAddress();
    }

    public AuthService(CachedUserDb userDb, CachedPageDb pageDb, Config appConfig, BlockingQueue<SendMailDto> sendMailQueue) {
        this.userDb = userDb;
        this.pageDb = pageDb;
        this.securityConfig = appConfig.get("security");
        this.sendMailQueue = sendMailQueue;
        this.adminMail = securityConfig.get("default.admin").asString().get();

        String keyConfig = securityConfig.get("key").asString().get();
        key = new SecretKeySpec(Base64.getDecoder().decode(keyConfig), SIGNATURE_ALGORITHM.getJcaName());

        tokenExpireHours = securityConfig.get("tokenExpireHours").asLong().get();

        createDefaults(securityConfig.get("default"));
    }

    @Override
    public void update(Rules rules) {
        rules
            .post("/login", Handler.create(CredentialDto.class, this::login))
            .get("/auth/{" + AUTH_REQUEST_PAGE_PARAM + "}", this::validateTokenCookie)
            .get("/auth", this::isLoggedIn)
            .put("/change-pw", Handler.create(ChangeCredentialDto.class, this::changePassword))
            .put("/activate", Handler.create(ChangeCredentialDto.class, this::activateUser));
    }

    public void login(ServerRequest request, ServerResponse response, CredentialDto credentialDto) {
        final String requestId = RequestLogHandler.parseRequestId(request);
        LOGGER.log(Level.FINE, "{0} trying login for {1}", new Object[]{requestId, credentialDto.mail() });

        UserEntity user = userDb.selectUserByMail(credentialDto.mail())
                                .orElseThrow(() -> new ForbiddenException(String.format("login denied for %s (login, mail not found)", credentialDto.mail()), "invalid credentials"));

        String hashedPassword = EncodingUtil.hashString(credentialDto.password(), user.getSalt());
        if(hashedPassword.equals(user.getPassword())) {
            LOGGER.log(Level.FINE, "{0} login granted for {1}", new Object[]{requestId, credentialDto.mail()});

            SecurityToken token = new SecurityToken(
                loadRemoteAddress(request),
                Date.from(Instant.now()),
                user.getId().toString(),
                user.getMail(),
                user.isAdmin(),
                user.isActive(),
                user.getPages()
                    .stream()
                    .filter(p -> p.isPrivatePage() && p.isHasAccess())
                    .map(UserPageEntity::getPageId)
                    .collect(Collectors.joining(SecurityToken.CLAIM_PRIVATE_ACCESS_DELIMITER)),
                Date.from(Instant.now().plus(tokenExpireHours, ChronoUnit.HOURS))
            );
            String jwtToken = createJwt(token);

            String domain = securityConfig.get("domain").asString().get();
            String expires = Date.from(Instant.now().plus(tokenExpireHours, ChronoUnit.HOURS)).toString();

            response.status(Status.OK_200);
            response.addHeader("Set-Cookie", String.format("%s=%s; Path=/; HttpOnly=true; "
                + "SameSite=Strict; Secure=true; Domain=%s; Expires=%s;",
                BEARER_COOKIE, jwtToken, domain, expires));
            response.send();

            sendMailQueue.add(new SendMailDto(
                Type.LOGIN_GRANT,
                token,
                adminMail
            ));
        } else
            throw new ForbiddenException(String.format("login denied for %s, (login, wrong pw)", credentialDto.mail()), "invalid credentials");
    }

    public void validateTokenCookie(ServerRequest request, ServerResponse response) {
        final String requestId = RequestLogHandler.parseRequestId(request);
        String pageIdParam = request.path().param(AUTH_REQUEST_PAGE_PARAM);

        if(pageIdParam == null || pageIdParam.isEmpty() || pageIdParam.isBlank())
            throw new BadRequestException(String.format("missing page param at auth request, from ip: %s",
                loadRemoteAddress(request)), "missing parameter");

        LOGGER.log(Level.FINE, "{0} validating token for request to page {1}", new Object[]{ requestId, pageIdParam });

        PageEntity page = pageDb.selectPage(pageIdParam)
                                .orElseThrow(() -> new ForbiddenException(String.format("access to unknown page denied, from: %s", loadRemoteAddress(request)), "not allowed"));

        if(!page.isPrivatePage()) {
            LOGGER.log(Level.FINE, "{0} access to public page granted: {1}",
                new Object[]{ requestId, page.getUrl() });
            response.status(Status.OK_200);
            response.send("enjoy!");
            return;
        }

        SecurityToken token = extractTokenCookie(request.headers());
        validateSecurityToken(request, token);
        if(token.getPrivatePagesAccess().contains(pageIdParam)) {
            LOGGER.log(Level.FINE, "{0} access to private page {1} granted for {2}",
                new Object[]{ requestId, page.getUrl(), token.getUserMail() });
            response.status(Status.OK_200);
            response.send("enjoy!");
        } else
            throw new ForbiddenException(String.format("access to private page %s refused for %s",
                page.getId(), token.getUserMail()), "not allowed");
    }

    public void isLoggedIn(ServerRequest request, ServerResponse response) {
        try {
            SecurityToken token = extractTokenCookie(request.headers());
            validateSecurityToken(request, token, true);
            UserEntity user = userDb.selectUser(UUID.fromString(token.getUserId()));

            if(user == null) {
                response.status(Status.NO_CONTENT_204);
                response.send();
                return;
            }

            response.send(new UserReadDto(user.getId(), user.getMail(), user.isAdmin(), user.isActive(), user.getPages()));
        } catch(SQLException e){
            throw new BadRequestException("unexpected db exception", "could not read user", e);
        }
    }

    public void changePassword(ServerRequest request, ServerResponse response, ChangeCredentialDto credentialDto) {
        changePassword(request, response, credentialDto, false);
    }

    private void changePassword(ServerRequest request, ServerResponse response, ChangeCredentialDto credentialDto, boolean fromActivation) {
        final String requestId = RequestLogHandler.parseRequestId(request);
        LOGGER.log(Level.FINE, "{0} changing password for {1}", new Object[]{ requestId, credentialDto.mail() });

        UserEntity user = userDb.selectUserByMail(credentialDto.mail())
                                .orElseThrow(() -> new ForbiddenException(String.format("login denied for %s (change password mail not found)", credentialDto.mail()), "invalid credentials"));
        if(!fromActivation) // if this method is called from activation, then the user won't be logged in
            verifyCurrentUserOrAdmin(request, user.getId());

        String oldHashedPassword = EncodingUtil.hashString(credentialDto.oldPassword(), user.getSalt());

        if(!oldHashedPassword.equals(user.getPassword()))
            throw new ForbiddenException(String.format("login denied for %s (change pw, old wrong)", credentialDto.mail()), "invalid credentials");

        String newHashedPassword = EncodingUtil.hashString(credentialDto.password(), user.getSalt());
        try {
            userDb.updateUserWithPages(user.getId(), newHashedPassword, user.isAdmin(), user.isActive(), new String[0], new String[0]);

            login(request, response, new CredentialDto(credentialDto.mail(), credentialDto.password()));
        } catch (SQLException e) {
            throw new BadRequestException("could not change password", "failed to change password", e);
        }
    }

    public void activateUser(ServerRequest request, ServerResponse response, ChangeCredentialDto credentialDto) {
        final String requestId = RequestLogHandler.parseRequestId(request);
        LOGGER.log(Level.FINE, "{0} activating user for {1}", new Object[]{ requestId, credentialDto.mail() });
        String activationTime = DateTimeFormatter.ofPattern("dd.MM.yyyy hh:mm:ss").format(LocalDateTime.now());

                                                                                   UserEntity user = userDb.selectUserByMail(credentialDto.mail())
                                .orElseThrow(() -> new ForbiddenException(String.format("login denied for %s (activate, mail not found)", credentialDto.mail()), "invalid credentials"));

        if(user.isActive())
            throw new BadRequestException("tried to activate already active user", "already active");

        try {
            userDb.activateUser(user.getId(), true);
            changePassword(request, response, credentialDto, true);
            sendMailQueue.add(new SendMailDto(Type.USER_ACTIVATED,
                                              new UserActivatedMailDto(user.getMail(), activationTime),
                                              adminMail));
        } catch (SQLException e) {
            throw new BadRequestException("unexpected db error", "activation failed, db error", e);
        }
    }

    public String createJwt(SecurityToken securityToken) {
        return Jwts.builder()
            .setIssuer(securityToken.getIssuer())
            .setIssuedAt(securityToken.getIssuedAt())
            .setExpiration(securityToken.getExpiresAt())
            .claim(SecurityToken.CLAIM_USER_ID, securityToken.getUserId())
            .claim(SecurityToken.CLAIM_USER_MAIL, securityToken.getUserMail())
            .claim(SecurityToken.CLAIM_ADMIN, securityToken.isAdmin())
            .claim(SecurityToken.CLAIM_ACTIVE, securityToken.isActive())
            .claim(SecurityToken.CLAIM_PRIVATE_ACCESS, securityToken.getPrivatePagesAccess())
            .signWith(SIGNATURE_ALGORITHM, key)
            .compact();
    }

    public SecurityToken parseJwt(String token) {
        SecurityToken securityToken = new SecurityToken();
        try {
            Claims claims = Jwts.parser().setSigningKey(key).parseClaimsJws(token).getBody();

            securityToken.setIssuer(claims.getIssuer());
            securityToken.setIssuedAt(claims.getIssuedAt());
            securityToken.setUserId(claims.get(SecurityToken.CLAIM_USER_ID, String.class));
            securityToken.setUserMail(claims.get(SecurityToken.CLAIM_USER_MAIL, String.class));
            securityToken.setAdmin(claims.get(SecurityToken.CLAIM_ADMIN, Boolean.class));
            securityToken.setActive(claims.get(SecurityToken.CLAIM_ACTIVE, Boolean.class));
            securityToken.setPrivatePagesAccess(claims.get(SecurityToken.CLAIM_PRIVATE_ACCESS, String.class));
            securityToken.setExpiresAt(claims.getExpiration());

            return securityToken;
        } catch (Exception e) {
            throw new UnauthorizedException(String.format("caught invalid token: %s - %s", e.getClass().getSimpleName(), e.getMessage()),
                "invalid token provided");
        }
    }

    public SecurityToken extractTokenCookie(RequestHeaders headers) {
        Optional<String> tokenCookie = headers.cookies().first(BEARER_COOKIE);

        if(tokenCookie.isEmpty())
            throw new UnauthorizedException("got admin request with no auth cookie", "unauthorized request");

        return parseJwt(tokenCookie.get());
    }

    private void verifyCurrentUserOrAdmin(ServerRequest request, UUID userId) {
        SecurityToken token = extractTokenCookie(request.headers());
        if(token.getUserId().equals(userId.toString()))
            return;

        if(!token.isAdmin())
            throw new ForbiddenException("user tried to modify other user and is not admin", "not allowed");
    }

    public void validateSecurityToken(ServerRequest request, SecurityToken token) {
        validateSecurityToken(request, token, false);
    }

    public void validateSecurityToken(ServerRequest request, SecurityToken token, boolean ignoreUserActive) {
        if(token == null)
            throw new UnauthorizedException(String.format("got request from %s, with no token provided",
                loadRemoteAddress(request)), "unauthorized");
        // handle client change (token rubbery dude)
        if(!token.getIssuer().equals(loadRemoteAddress(request))) {
            throw new UnauthorizedException(String.format("invalid issuer in request: %s changed to %s, associated user: %s",
                token.getIssuer(), loadRemoteAddress(request), token.getUserMail()), "unauthorized request");
            // handle token expired
        } else if(token.getExpiresAt().before(Date.from(Instant.now()))) {
            throw new UnauthorizedException(String.format("token for %s expired",
                token.getUserMail()), "token expired");
        } else if(!ignoreUserActive && !token.isActive()) {
            throw new InactiveUserException(token.getUserMail());
        }
    }

    private void createDefaults(Config defaultConfig) {
        String defaultMail = defaultConfig.get("mail").as(String.class).get();
        String defaultPassword = defaultConfig.get("password").as(String.class).get();

        createDefaultAdmin(defaultMail, defaultPassword);
    }

    private void createDefaultAdmin(String defaultMail, String defaultPassword) {
        LOGGER.log(Level.FINE, "checking if default user with mail {0} exists", defaultMail);

        Optional<UserEntity> existingDefaultUser = userDb.selectUserByMail(defaultMail);
        if(existingDefaultUser.isEmpty()) {
            LOGGER.log(Level.FINE, "default user not found -> creating one");

            String salt = EncodingUtil.generateSalt();
            String hashedPassword = EncodingUtil.hashString(defaultPassword, salt);

            try {
                userDb.insertUser(UUID.randomUUID(), defaultMail, hashedPassword, salt, true, true, new String[0]);
            } catch (SQLException e) {
                LOGGER.log(Level.WARNING, "could not create default admin", e);
            }
        }
    }
}
