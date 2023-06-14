/*
 * -----------------------------------------------------------------------------
 * © Swisslog AG
 * Swisslog is not liable for any usage of this source code that is not agreed on between Swisslog and the other party.
 * The mandatory legal liability remains unaffected.
 * -----------------------------------------------------------------------------
 */

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
    private Optional<CityResponse> location;
}
