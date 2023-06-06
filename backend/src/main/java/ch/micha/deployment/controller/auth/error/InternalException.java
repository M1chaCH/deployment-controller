package ch.micha.deployment.controller.auth.error;

import io.helidon.common.http.Http.Status;

public class InternalException extends AppRequestException{

    public InternalException(String serverMessage, String clientMessage) {
        super(serverMessage, clientMessage, Status.INTERNAL_SERVER_ERROR_500, true);
    }

    public InternalException(String serverMessage, String clientMessage, Throwable cause) {
        super(serverMessage, clientMessage, Status.INTERNAL_SERVER_ERROR_500, cause, true);
    }
}
