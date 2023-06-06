package ch.micha.deployment.controller.auth.error;

import io.helidon.common.http.Http.Status;

public class BadRequestException extends AppRequestException{

    public BadRequestException(String serverMessage, String clientMessage) {
        super(serverMessage, clientMessage, Status.BAD_REQUEST_400, false);
    }

    public BadRequestException(String serverMessage, String clientMessage, Throwable cause) {
        super(serverMessage, clientMessage, Status.BAD_REQUEST_400, cause, false);
    }
}
