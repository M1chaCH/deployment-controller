package ch.micha.deployment.controller.auth.location;

import com.maxmind.geoip2.model.CityResponse;

import java.time.Instant;
import java.time.temporal.ChronoUnit;
import java.util.*;
import java.util.logging.Level;
import java.util.logging.Logger;

public class LocationCache {
    private static final Logger LOGGER = Logger.getLogger(LocationCache.class.getSimpleName());
    private final Map<String, CityResponse> cachedCities = new HashMap<>();
    private final Set<String> failedIps = new HashSet<>();
    private final int hoursLifetime;
    private Date expiresAt;

    public LocationCache(int hoursLifetime) {
        this.hoursLifetime = hoursLifetime;
        handleExpired();
    }

    public Optional<CityResponse> get(String remoteAddress) {
        handleExpired();
        CityResponse cachedCity = cachedCities.get(remoteAddress);
        return cachedCity == null ? Optional.empty() : Optional.of(cachedCity);
    }

    public void put(String remoteAddress, CityResponse city){
        cachedCities.put(remoteAddress, city);
    }

    public boolean hasFailed(String remoteAddress) {
        handleExpired();
        return failedIps.stream().anyMatch(ip -> ip.equals(remoteAddress));
    }

    public void putFailed(String remoteAddress){
        failedIps.add(remoteAddress);
    }

    private void handleExpired() {
        if(expiresAt == null || expiresAt.before(Date.from(Instant.now()))) {
            cachedCities.clear();
            failedIps.clear();
            expiresAt = Date.from(Instant.now().plus(hoursLifetime, ChronoUnit.HOURS));
            LOGGER.log(Level.FINE, "reset location cache, expires in {0}h", new Object[]{ hoursLifetime });
        }
    }
}
