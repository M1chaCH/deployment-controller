package ch.micha.deployment.controller.auth.error;

import io.helidon.common.http.Http.Status;

public class NotFoundException extends AppRequestException{

    public NotFoundException(String serverMessage, String clientMessage) {
        super(serverMessage, clientMessage, Status.NOT_FOUND_404, false);
    }

    public NotFoundException(String serverMessage, String clientMessage, Throwable cause) {
        super(serverMessage, clientMessage, Status.NOT_FOUND_404, cause, false);
    }
}
