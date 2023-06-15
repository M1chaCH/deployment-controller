package ch.micha.deployment.controller.auth.logging;

import ch.micha.deployment.controller.auth.location.LocationResolver;
import com.maxmind.geoip2.model.CityResponse;
import io.helidon.config.Config;

import java.io.IOException;
import java.time.Instant;
import java.time.ZoneId;
import java.time.format.DateTimeFormatter;
import java.util.Optional;
import java.util.concurrent.BlockingQueue;
import java.util.logging.Level;
import java.util.logging.Logger;

public class RequestLogProcessor implements Runnable{
    private static final Logger LOGGER = Logger.getLogger(RequestLogProcessor.class.getSimpleName());
    private static final DateTimeFormatter DATE_TIME_FORMATTER = DateTimeFormatter.ofPattern("dd.MM.yyyy HH:mm:ss");

    private final BlockingQueue<RequestLogDto> requestLogQueue;
    private final LocationResolver locationResolver;
    private final RequestLogFileWriter writer;


    public RequestLogProcessor(BlockingQueue<RequestLogDto> requestLogQueue, String absoluteLogDir, Config locationConfig) throws IOException {
        this.requestLogQueue = requestLogQueue;

        LOGGER.log(Level.INFO, "initializing request log processor", new Object[]{ });
        this.locationResolver = LocationResolver.getInstance(locationConfig);
        writer = new RequestLogFileWriter(absoluteLogDir);
    }

    @SuppressWarnings({"java:S2189", "InfiniteLoopStatement"}) // it makes sense for this loop to be infinite
    @Override
    public void run() {
        LOGGER.log(Level.INFO, "request log processor thread created & started: {0}", new Object[]{ Thread.currentThread().getName() });

        try {
            while (true) {
                LOGGER.log(Level.FINE, "waiting for request");
                RequestLogDto currentRequest = requestLogQueue.take();
                LOGGER.log(Level.FINE, "{0} processing request {1}", new Object[]{ currentRequest.getId(), currentRequest.getRemoteAddress() });

                currentRequest.setLocation(locationResolver.resolveLocation(currentRequest.getRemoteAddress()));
                String country = "unknown";
                String city = "unknown";
                Optional<CityResponse> location = currentRequest.getLocation();
                if(location.isPresent()) {
                    country = location.get().getCountry().getName();
                    city = location.get().getCity().getName();
                }
                String message = String.format("%s %s: %s %s | %s - %s, %s -> %s %s %sms",
                        formatInstant(currentRequest.getRequestStart()),
                        currentRequest.getId(),
                        currentRequest.getMethod().name(),
                        currentRequest.getRequestPath(),
                        currentRequest.getRemoteAddress(),
                        country,
                        city,
                        currentRequest.getStatus().code(),
                        currentRequest.getStatus().reasonPhrase(),
                        currentRequest.getDuration());
                writer.writeLine(message);
            }
        } catch (InterruptedException e){
            LOGGER.log(Level.WARNING, "{0} interrupted -> re-interrupting", new Object[]{ Thread.currentThread().getName() });
            Thread.currentThread().interrupt();
        } catch (IOException e) {
            LOGGER.log(Level.WARNING, "IOException while writing to access log, {0}", new Object[]{ e.getMessage() });
        }

    }

    private String formatInstant(Instant time) {
        return time.atZone(ZoneId.systemDefault()).format(DATE_TIME_FORMATTER);
    }
}
