package ch.micha.deployment.controller.auth.location;

import com.maxmind.geoip2.WebServiceClient;
import com.maxmind.geoip2.exception.GeoIp2Exception;
import com.maxmind.geoip2.model.CityResponse;
import io.helidon.config.Config;

import java.io.IOException;
import java.net.InetAddress;
import java.util.Optional;
import java.util.logging.Level;
import java.util.logging.Logger;

public class LocationResolver {
    private static final Logger LOGGER = Logger.getLogger(LocationResolver.class.getSimpleName());
    private static final String LOCAL_REMOTE_ADDRESS = "0:0:0:0:0:0:0:1";
    private static LocationResolver instant;

    private final WebServiceClient locationWebClient;
    private final LocationCache locationCache;

    public static synchronized LocationResolver getInstance(Config locationConfig) {
        if(instant == null)
            instant = new LocationResolver(locationConfig);
        return instant;
    }

    private LocationResolver(Config locationConfig) {
        locationWebClient = initializeLocationWebClient(locationConfig);
        locationCache = new LocationCache(locationConfig.get("cacheExpireHours").as(Integer.class).get());
    }

    public synchronized Optional<CityResponse> resolveLocation(String remoteAddress) {
        Optional<CityResponse> cachedCity = locationCache.get(remoteAddress);
        if(cachedCity.isPresent()) {
            LOGGER.log(Level.FINE, "resolved (cached!) location for {0}: {1} -> {2}",
                    new Object[]{ remoteAddress, cachedCity.get().getCountry().getName(), cachedCity.get().getCity().getName() });
            return cachedCity;
        } else if(locationCache.hasFailed(remoteAddress)){
            LOGGER.log(Level.FINE, "resolved (cached!) location for {0}: not found!",
                new Object[]{ remoteAddress });
            return Optional.empty();
        }

        try{
            CityResponse requestLocation;
            if(LOCAL_REMOTE_ADDRESS.equals(remoteAddress))
                requestLocation = locationWebClient.city();
            else
                requestLocation = locationWebClient.city(InetAddress.getByName(remoteAddress));

            LOGGER.log(Level.FINE, "resolved location for {0}: {1} -> {2}",
                new Object[]{ remoteAddress, requestLocation.getCountry().getName(), requestLocation.getCity().getName() });
            locationCache.put(remoteAddress, requestLocation);
            return Optional.of(requestLocation);
        } catch (GeoIp2Exception e) {
            if(e.getMessage().contains("is a reserved IP address"))
                LOGGER.log(Level.WARNING, "failed to load location from request: {0}, request came from within the network",
                    new Object[]{ remoteAddress });
            else
                LOGGER.log(Level.SEVERE, "failed to load location from request: {0} - {1}",
                    new Object[]{ remoteAddress, e.getMessage() });
        } catch (IOException ioe) {
            LOGGER.log(Level.SEVERE, "ioexception during location loading: {0} - {1}",
                new Object[]{ remoteAddress, ioe.getMessage() });
        }
        locationCache.putFailed(remoteAddress);
        return Optional.empty();
    }

    private WebServiceClient initializeLocationWebClient(Config locationConfig) {
        String host = locationConfig.get("host").asString().get();
        Integer account = locationConfig.get("account").as(Integer.class).get();
        String license = locationConfig.get("license").asString().get();

        return new WebServiceClient.Builder(account, license).host(host).build();
    }
}
