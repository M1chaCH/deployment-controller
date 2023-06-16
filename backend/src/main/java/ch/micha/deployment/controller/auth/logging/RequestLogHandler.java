package ch.micha.deployment.controller.auth.logging;

import ch.micha.deployment.controller.auth.auth.AuthService;
import io.helidon.config.Config;
import io.helidon.webserver.Handler;
import io.helidon.webserver.ServerRequest;
import io.helidon.webserver.ServerResponse;

import java.io.IOException;
import java.time.Instant;
import java.util.Optional;
import java.util.Random;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.LinkedBlockingQueue;
import java.util.logging.Level;
import java.util.logging.Logger;

public class RequestLogHandler implements Handler {
    public static final String REQUEST_START_CONTEXT = "START";
    public static final String REQUEST_ID_CONTEXT = "ID";

    private static final Logger LOGGER = Logger.getLogger(RequestLogHandler.class.getSimpleName());

    private final Random random = new Random();
    private final BlockingQueue<RequestLogDto> requestLogQueue = new LinkedBlockingQueue<>();

    public static String parseRequestId(ServerRequest request) {
        return request.context().get(REQUEST_ID_CONTEXT, String.class).orElse("unknown!");
    }

    public RequestLogHandler(Config appConfig) {
        try {
            Thread requestLogProcessor = new Thread(
                    new RequestLogProcessor(requestLogQueue, appConfig.get("logs").get("directory").asString().get(),
                            appConfig.get("location"))
            );
            requestLogProcessor.setName("request-log-processor");
            requestLogProcessor.start();
        } catch (IOException e){
            LOGGER.log(Level.WARNING, "failed to create request log processor, wont process logs!!", e);
        }
    }

    @Override
    public void accept(ServerRequest request, ServerResponse response) {
        final Instant requestStart = Instant.now();
        request.context().register(REQUEST_START_CONTEXT, requestStart);
        final String requestId = String.valueOf(random.nextInt(900000) + 100000);
        request.context().register(REQUEST_ID_CONTEXT, requestId);

        LOGGER.log(Level.INFO, "request incoming {0}: {1} - {2} | {3}", new Object[]{
                requestId,
                request.method().name(),
                request.requestedUri().path(),
                AuthService.loadRemoteAddress(request)
        });

        response.whenSent().thenAcceptAsync(sentResponse -> {
            long duration = Instant.now().toEpochMilli() - requestStart.toEpochMilli();
            LOGGER.log(Level.INFO, "request done {0}: {1} - {2} | {3}ms {4} {5}", new Object[]{
                    requestId,
                    request.method().name(),
                    request.requestedUri().path(),
                    duration,
                    sentResponse.status().code(),
                    sentResponse.status().reasonPhrase()
            });
            requestLogQueue.add(new RequestLogDto(
                    requestId,
                    request.method(),
                    AuthService.loadRemoteAddress(request),
                    request.requestedUri().path(),
                    sentResponse.status(),
                    requestStart,
                    duration,
                    Optional.empty()
            ));
        });

        request.next();
    }
}
