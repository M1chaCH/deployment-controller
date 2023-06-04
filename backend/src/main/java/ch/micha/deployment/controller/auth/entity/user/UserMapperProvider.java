package ch.micha.deployment.controller.auth.entity.user;

import ch.micha.deployment.controller.auth.entity.adduser.AddUser;
import ch.micha.deployment.controller.auth.entity.adduser.AddUserMapper;
import ch.micha.deployment.controller.auth.entity.edituser.EditUserMapper;
import io.helidon.dbclient.DbMapper;
import io.helidon.dbclient.spi.DbMapperProvider;
import jakarta.annotation.Priority;
import java.util.Optional;

@Priority(1000)
public class UserMapperProvider implements DbMapperProvider {
    private static final UserMapper USER_MAPPER = new UserMapper();
    private static final AddUserMapper ADD_USER_MAPPER = new AddUserMapper();
    private static final EditUserMapper EDIT_USER_MAPPER = new EditUserMapper();

    @SuppressWarnings("unchecked")
    @Override
    public <T> Optional<DbMapper<T>> mapper(Class<T> type) {
        if(type.equals(User.class))
            return Optional.of((DbMapper<T>) USER_MAPPER);
        if(type.equals(AddUser.class))
            return Optional.of((DbMapper<T>) ADD_USER_MAPPER);
        if(type.equals(EditUserMapper.class))
            return Optional.of((DbMapper<T>) EDIT_USER_MAPPER);

        return Optional.empty();
    }
}
