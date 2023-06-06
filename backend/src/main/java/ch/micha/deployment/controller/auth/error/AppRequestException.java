package ch.micha.deployment.controller.auth.error;

import io.helidon.common.http.Http.Status;
import io.helidon.webserver.ServerResponse;
import java.util.concurrent.CompletionException;
import java.util.logging.Level;
import java.util.logging.Logger;
import lombok.Getter;
import org.postgresql.util.PSQLException;

@Getter
public abstract class AppRequestException extends RuntimeException{
    private static final Logger LOGGER = Logger.getLogger(AppRequestException.class.getSimpleName());

    private final Status httpStatus;
    private final String serverMessage;
    private final String clientMessage;
    private final boolean unexpected;

    public static AppRequestException fittingException (Throwable throwable) {
        String baseServerMessage = "caught exception:";

        Throwable realCause = throwable;
        if (throwable instanceof CompletionException)
            realCause = throwable.getCause();

        if(realCause instanceof PSQLException psqlException &&
            psqlException.getMessage().contains("value violates unique constraint"))
            return new BadRequestException(
                String.format("%s unique constraint was violated: %s",
                    baseServerMessage, psqlException.getServerErrorMessage()),
                "already exists", psqlException);

        return new InternalException("caught unknown error", "unexpected error, please retry later", realCause);
    }

    public static Void respondFitting(ServerResponse response, Throwable cause) {
        AppRequestException.fittingException(cause).sendResponse(response);
        return null;
    }

    protected AppRequestException(String serverMessage, String clientMessage, Status httpStatus, boolean unexpected) {
        super(serverMessage);
        this.httpStatus = httpStatus;
        this.serverMessage = serverMessage;
        this.clientMessage = clientMessage;
        this.unexpected = unexpected;
    }

    protected AppRequestException(String serverMessage, String clientMessage, Status httpStatus, Throwable cause, boolean unexpected) {
        super(serverMessage, cause);
        this.httpStatus = httpStatus;
        this.serverMessage = serverMessage;
        this.clientMessage = clientMessage;
        this.unexpected = unexpected;
    }

    public void sendResponse(ServerResponse response) {
        if(isUnexpected())
            LOGGER.log(Level.SEVERE, "caught unexpected error", this);
        else
            LOGGER.log(Level.INFO, "handling error {0}: {1}, responding with {2}",
                new Object[]{ getClass().getSimpleName(), getServerMessage(), getHttpStatus().code()});

        response.status(getHttpStatus());
        response.send(getClientMessage());
    }
}
