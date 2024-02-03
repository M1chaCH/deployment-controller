package ch.micha.deployment.controller.auth.db;

import ch.micha.deployment.controller.auth.error.NotFoundException;
import java.sql.Connection;
import java.sql.PreparedStatement;
import java.sql.ResultSet;
import java.sql.SQLException;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.UUID;
import java.util.logging.Level;
import java.util.logging.Logger;

public class CachedUserDb {
    private static final Logger LOGGER = Logger.getLogger(CachedUserDb.class.getSimpleName());

    protected final Connection db;
    protected final UserPageDb userPageDb;
    protected final Map<UUID, UserEntity> cache = new HashMap<>();

    public CachedUserDb(Connection db, UserPageDb userPageDb) {
        this.db = db;
        this.userPageDb = userPageDb;
    }

    public UserEntity selectUser(UUID userId) throws SQLException {
        UserEntity selectedUser = cache.get(userId);
        if(selectedUser != null)
            return selectedUser;

        PreparedStatement userStatement = db.prepareStatement("""
            select *
            from users as u
            where u.id = ?
            """);
        userStatement.setObject(1, userId);
        ResultSet userResult = userStatement.executeQuery();

        if(!userResult.next()) {
            throw new NotFoundException("could not select user, fetch size was not 1", "could not find user");
        }

        selectedUser = parseUserResult(userResult);
        userResult.close();
        userStatement.close();

        selectedUser.addPages(userPageDb.selectPagesForUser(userId));
        cache.put(userId, selectedUser);
        return selectedUser;
    }

    public Optional<UserEntity> selectUserByMail(String mail) {
        Optional<UserEntity> cachedUser = cache.values()
                    .stream()
                    .filter(u -> u.getMail().equals(mail))
                    .findAny();

        if(cachedUser.isPresent())
            return cachedUser;

        try {
            PreparedStatement userStatement = db.prepareStatement("""
            select *
            from users as u
            where u.mail = ?
            """);
            userStatement.setString(1, mail);
            ResultSet userResult = userStatement.executeQuery();

            if(!userResult.next()) {
                return Optional.empty();
            }

            UserEntity selectedUser = parseUserResult(userResult);
            userResult.close();
            userStatement.close();

            selectedUser.addPages(userPageDb.selectPagesForUser(selectedUser.getId()));
            cache.put(selectedUser.getId(), selectedUser);
            return Optional.of(selectedUser);
        } catch (SQLException e) {
            LOGGER.log(Level.WARNING, "could not read user by mail, %s".formatted(mail), e);
            return Optional.empty();
        }
    }

    public List<UserEntity> selectUsers() throws SQLException {
        PreparedStatement usersStatement = db.prepareStatement("""
            select * from users order by users.mail
            """);
        ResultSet usersResult = usersStatement.executeQuery();

        List<UserEntity> users = new ArrayList<>();
        while(usersResult.next()) {
            UserEntity user = parseUserResult(usersResult);
            user.addPages(userPageDb.selectPagesForUser(user.getId()));
            users.add(user);
            cache.put(user.getId(), user);
        }

        usersResult.close();
        usersStatement.close();
        return users;
    }

    protected UserEntity parseUserResult(ResultSet userResult) throws SQLException {
        return new UserEntity(
            userResult.getObject("id", UUID.class),
            userResult.getString("mail"),
            userResult.getString("password"),
            userResult.getString("salt"),
            userResult.getBoolean("admin"),
            userResult.getBoolean("active"),
            userResult.getTimestamp("created_at").toLocalDateTime(),
            userResult.getTimestamp("last_login").toLocalDateTime()
        );
    }

    public void insertUser(UUID userId, String mail, String password, String salt, boolean admin, String[] pageAccess) throws SQLException {
        insertUser(userId, mail, password, salt, admin, false, pageAccess);
    }

    public void insertUser(UUID userId, String mail, String password, String salt, boolean admin, boolean active, String[] pageAccess) throws SQLException {
        PreparedStatement statement = db.prepareStatement("""
            insert into users (id, mail, password, salt, admin, active)
            values (?, ?, ?, ?, ?, ?)
            """);
        statement.setObject(1, userId);
        statement.setString(2, mail);
        statement.setString(3, password);
        statement.setString(4, salt);
        statement.setBoolean(5, admin);
        statement.setBoolean(6, active);
        statement.execute();
        statement.close();

        userPageDb.insertPagesForUser(userId, pageAccess);

        selectUser(userId); // just to update cache
    }

    public void updateUserWithPages(UUID userId, String password, boolean admin, boolean active, String[] addPageAccess, String[] deletePageAccess) throws SQLException {
        PreparedStatement statement = db.prepareStatement("""
            update users set password = ?, admin = ?, active = ?
            where id = ?
            """);
        statement.setString(1, password);
        statement.setBoolean(2, admin);
        statement.setBoolean(3, active);
        statement.setObject(4, userId);

        statement.executeUpdate();
        statement.close();

        userPageDb.insertPagesForUser(userId, addPageAccess);
        userPageDb.deletePagesForUser(userId, deletePageAccess);

        cache.remove(userId);
        selectUser(userId); // just to update cache
    }

    public void activateUser(UUID userId, boolean active) throws SQLException {
        PreparedStatement statement = db.prepareStatement("""
            update users set active = ?
            where id = ?
            """);
        statement.setBoolean(1, active);
        statement.setObject(2, userId);

        statement.executeUpdate();
        statement.close();

        cache.remove(userId);
        selectUser(userId); // just to update cache
    }

    public long countAdminUsers() {
        try {
            PreparedStatement countStatement = db.prepareStatement("""
                select count(id) as count
                from users
                where users.admin = true
                """);

            ResultSet result = countStatement.executeQuery();
            result.next();
            long count = result.getLong("count");

            result.close();
            countStatement.close();

            return count;
        } catch (SQLException e) {
            return -1L;
        }
    }

    public void deleteUser(UUID userId) throws SQLException {
        PreparedStatement statement = db.prepareStatement("""
            delete from users where id = ?
            """);
        statement.setObject(1, userId);
        statement.execute();
        statement.close();

        cache.remove(userId);
    }

    public void invalidateCache() {
        cache.clear();
    }
}
