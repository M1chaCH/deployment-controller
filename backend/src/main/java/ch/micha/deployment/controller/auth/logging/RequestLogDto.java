package ch.micha.deployment.controller.auth.logging;

import com.maxmind.geoip2.model.CityResponse;
import io.helidon.common.http.Http;
import lombok.AllArgsConstructor;
import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.Setter;

import java.time.Instant;
import java.util.Optional;

@AllArgsConstructor
@NoArgsConstructor
@Getter
@Setter
public class RequestLogDto {
    private String id;
    private Http.RequestMethod method;
    private String remoteAddress;
    private String requestPath;
    private Http.ResponseStatus status;
    private Instant requestStart;
    private long duration;
    @SuppressWarnings("OptionalUsedAsFieldOrParameterType") // intended to be an optional field
    private Optional<CityResponse> location;
}
