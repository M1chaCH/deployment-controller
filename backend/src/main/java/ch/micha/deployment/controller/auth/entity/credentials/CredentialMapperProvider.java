package ch.micha.deployment.controller.auth.entity.credentials;

import ch.micha.deployment.controller.auth.entity.user.User;
import io.helidon.dbclient.DbMapper;
import io.helidon.dbclient.spi.DbMapperProvider;
import jakarta.annotation.Priority;
import java.util.Optional;

@Priority(1000)
public class CredentialMapperProvider implements DbMapperProvider {
    private static final CredentialMapper CREDENTIAL_MAPPER = new CredentialMapper();

    @SuppressWarnings("unchecked")
    @Override
    public <T> Optional<DbMapper<T>> mapper(Class<T> type) {
        if(type.equals(User.class))
            return Optional.of((DbMapper<T>) CREDENTIAL_MAPPER);

        return Optional.empty();
    }
}
