package ch.micha.deployment.controller.auth.auth;

import io.helidon.common.http.Http.Status;
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
        SecurityToken token = service.extractTokenCookie(request.headers(), response);
        if(token == null)
            return; // extract token responds with an error -> we don't have to do anything here

        service.validateSecurityToken(request, token);
        if(!token.isAdmin()) {
            LOGGER.log(Level.WARNING, "{0} tried to access admin",
                new Object[]{ token.getUserMail() });
            response.status(Status.FORBIDDEN_403);
            response.send("not allowed");
            return;
        }

        LOGGER.log(Level.INFO, "authorized admin request to {0}", new Object[]{ token.getUserMail() });
        request.next();
    }
}
