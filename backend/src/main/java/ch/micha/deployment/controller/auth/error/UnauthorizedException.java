package ch.micha.deployment.controller.auth.error;

import io.helidon.common.http.Http.Status;

public class UnauthorizedException extends AppRequestException{

    public UnauthorizedException(String serverMessage, String clientMessage) {
        super(serverMessage, clientMessage, Status.UNAUTHORIZED_401, false);
    }

    public UnauthorizedException(String serverMessage, String clientMessage, Throwable cause) {
        super(serverMessage, clientMessage, Status.UNAUTHORIZED_401, cause, false);
    }
}
