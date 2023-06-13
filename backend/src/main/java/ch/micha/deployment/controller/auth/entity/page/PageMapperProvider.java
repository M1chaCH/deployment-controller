package ch.micha.deployment.controller.auth.entity.page;

import io.helidon.dbclient.DbMapper;
import io.helidon.dbclient.spi.DbMapperProvider;
import jakarta.annotation.Priority;
import java.util.Optional;

@Priority(1000)
public class PageMapperProvider implements DbMapperProvider {
    private static final PageMapper PAGE_MAPPER = new PageMapper();

    @SuppressWarnings("unchecked")
    @Override
    public <T> Optional<DbMapper<T>> mapper(Class<T> type) {
        if(type.equals(Page.class))
            return Optional.of((DbMapper<T>) PAGE_MAPPER);

        return Optional.empty();
    }
}
