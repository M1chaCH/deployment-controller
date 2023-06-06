package ch.micha.deployment.controller.auth.error;

import io.helidon.common.http.Http.Status;

public class ForbiddenException extends AppRequestException{

    public ForbiddenException(String serverMessage, String clientMessage) {
        super(serverMessage, clientMessage, Status.FORBIDDEN_403, false);
    }

    public ForbiddenException(String serverMessage, String clientMessage, Throwable cause) {
        super(serverMessage, clientMessage, Status.FORBIDDEN_403, cause, false);
    }
}
