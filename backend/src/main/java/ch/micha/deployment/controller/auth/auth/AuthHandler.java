package ch.micha.deployment.controller.auth.auth;

import ch.micha.deployment.controller.auth.error.ForbiddenException;
import ch.micha.deployment.controller.auth.logging.RequestLogHandler;
import io.helidon.common.http.Http.Method;
import io.helidon.webserver.Handler;
import io.helidon.webserver.ServerRequest;
import io.helidon.webserver.ServerResponse;
import java.util.logging.Level;
import java.util.logging.Logger;

public class AuthHandler implements Handler {
    private static final Logger LOGGER = Logger.getLogger(AuthHandler.class.getSimpleName());

    private final AuthService service;

    public AuthHandler(AuthService authService) {
        this.service = authService;
    }

    @Override
    public void accept(ServerRequest request, ServerResponse response) {
        String requestId = RequestLogHandler.parseRequestId(request);

        if(request.method().equals(Method.GET) && request.requestedUri().path().equals("/pages")) {
            LOGGER.log(Level.FINE, "{0} GET request to /pages -> ignoring auth", requestId);
            request.next();
            return;
        }

        SecurityToken token = service.extractTokenCookie(request.headers());
        if(token == null)
            return; // extract token responds with an error -> we don't have to do anything here

        service.validateSecurityToken(request, token);
        if(!token.isAdmin())
            throw new ForbiddenException(String.format("%s tried to access admin", token.getUserMail()), "not allowed");

        LOGGER.log(Level.FINE, "{0} authorized admin request to {1}", new Object[]{ requestId, token.getUserMail() });
        request.next();
    }
}
