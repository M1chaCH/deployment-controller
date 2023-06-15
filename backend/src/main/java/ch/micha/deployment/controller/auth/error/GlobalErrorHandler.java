package ch.micha.deployment.controller.auth.error;

import ch.micha.deployment.controller.auth.logging.RequestLogHandler;
import io.helidon.common.http.Http;
import io.helidon.webserver.HttpException;
import io.helidon.webserver.ServerRequest;
import io.helidon.webserver.ServerResponse;

import java.util.logging.Level;
import java.util.logging.Logger;

public class GlobalErrorHandler {
    public static final Logger LOGGER = Logger.getLogger(GlobalErrorHandler.class.getSimpleName());

    public void handleAppRequestException(ServerRequest request, ServerResponse response, AppRequestException exception) {
        exception.sendResponse(response, RequestLogHandler.parseRequestId(request));
    }

    public void handleHttpException(ServerRequest request, ServerResponse response, HttpException exception) {
        String requestId = RequestLogHandler.parseRequestId(request);
        LOGGER.log(Level.WARNING, "{0} caught http exception: {1}, responding: {2} {3}",
            new Object[]{ requestId, exception.getMessage(), exception.status().code(), exception.status().reasonPhrase() });

        response.status(exception.status().code());
        response.send(exception.getMessage());
    }

    public void handleException(ServerRequest request, ServerResponse response, Exception exception) {
        String requestId = RequestLogHandler.parseRequestId(request);
        LOGGER.log(Level.SEVERE, String.format("%s unexpected error, sending bad request", requestId), exception);

        response.status(Http.Status.BAD_REQUEST_400);
        response.send(String.format("error: %s", exception.getClass().getSimpleName()));
    }
}
