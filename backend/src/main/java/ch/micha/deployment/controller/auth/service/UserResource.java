package ch.micha.deployment.controller.auth.service;

import ch.micha.deployment.controller.auth.EncodingUtil;
import ch.micha.deployment.controller.auth.entity.adduser.AddUser;
import ch.micha.deployment.controller.auth.entity.edituser.EditUser;
import ch.micha.deployment.controller.auth.entity.user.User;
import io.helidon.common.http.Http;
import io.helidon.common.http.Http.Status;
import io.helidon.dbclient.DbClient;
import io.helidon.dbclient.DbRow;
import io.helidon.webserver.Handler;
import io.helidon.webserver.Routing.Rules;
import io.helidon.webserver.ServerRequest;
import io.helidon.webserver.ServerResponse;
import io.helidon.webserver.Service;
import jakarta.json.JsonObject;
import java.util.Optional;
import java.util.concurrent.CompletionException;
import java.util.logging.Level;
import java.util.logging.Logger;
import org.postgresql.util.PSQLException;

public class UserResource implements Service{
    private static final Logger LOGGER = Logger.getLogger(UserResource.class.getSimpleName());

    private final DbClient db;

    public UserResource(DbClient db) {
        this.db = db;
    }

    @Override
    public void update(Rules rules) {
        rules
            .get("/", this::getUsers)
            .post("/", Handler.create(AddUser.class, this::addUser))
            .put("/", Handler.create(EditUser.class, this::editUser))
            .delete("/{id}", this::deleteUser);
    }

    private void getUsers(ServerRequest request, ServerResponse response) {
        LOGGER.info("loading all users");
        response.send(db.execute(exec -> exec.namedQuery("select-users"))
                .map(item -> item.as(JsonObject.class)), JsonObject.class)
            .thenAccept(sentResponse ->
                LOGGER.log(Level.INFO, "{0} - successfully loaded all users", sentResponse.status()));
    }

    private void addUser(ServerRequest request, ServerResponse response, AddUser toAdd) {
        LOGGER.log(Level.INFO, "adding user {0} as admin:{1}", new Object[]{ toAdd.mail(), toAdd.admin() });

        String salt = EncodingUtil.generateSalt();
        String hashedPassword = EncodingUtil.hashString(toAdd.password(), salt);
        final User user = new User(-1, toAdd.mail(), hashedPassword, salt, toAdd.admin(),
            null, null);

        db.execute(exec -> exec
                .createNamedInsert("insert-user")
                .namedParam(user)
                .execute())
            .thenAccept(count -> {
                LOGGER.log(Level.INFO, "added {0} user(s)", count);
                response.status(Status.NO_CONTENT_204);
                response.send();
            })
            .exceptionally(throwable -> sendError(throwable, response));
    }

    private void editUser(ServerRequest request, ServerResponse response, EditUser toEdit) {
        LOGGER.log(Level.INFO, "updating user {0}-{1}", new Object[]{ toEdit.id(), toEdit.mail() });

        Optional<DbRow> userRow = db.execute(exec -> exec
            .createNamedGet("select-user")
            .addParam("id", toEdit.id())
            .execute()).await();

        if(userRow.isEmpty()) {
            sendNotFound(response, String.valueOf(toEdit.id()));
            return;
        }

        User user = userRow.get().as(User.class);
        String newHashedPassword = EncodingUtil.hashString(toEdit.password(), user.salt());
        db.execute(exec -> exec
                .createNamedUpdate("update-user")
                .addParam("id", toEdit.id())
                .addParam("mail", toEdit.mail())
                .addParam("password", newHashedPassword)
                .addParam("admin", toEdit.admin())
                .execute())
            .thenAccept(count -> {
                LOGGER.log(Level.INFO, "changed {0} user(s)", new Object[]{ count });
                response.status(Status.NO_CONTENT_204);
                response.send();
            })
            .exceptionally(t -> sendError(t, response));
    }

    private void deleteUser(ServerRequest request, ServerResponse response) {
        int userId = Integer.parseInt(request.path().param("id"));

        LOGGER.log(Level.INFO, "deleting user with id {0}", new Object[]{ userId });

        db.execute(exec -> exec
                .createNamedDelete("delete-user")
                .addParam("id", userId)
                .execute())
            .thenAccept(count -> {
                LOGGER.log(Level.INFO, "deleted {0} user(s)", new Object[]{ count });
                response.status(Status.NO_CONTENT_204);
                response.send();
            })
            .exceptionally(t -> sendError(t, response));
    }

    private void sendNotFound(ServerResponse response, String id) {
        response.status(Status.NOT_FOUND_404);
        response.send(String.format("Could not find user by %s", id));
    }

    @SuppressWarnings({"java:S3516", "SameReturnValue"})
    private <T> T sendError(Throwable throwable, ServerResponse response) {
        Throwable realCause = throwable;
        if (throwable instanceof CompletionException)
            realCause = throwable.getCause();

        if(realCause instanceof PSQLException psqlException &&
            psqlException.getMessage().contains("value violates unique constraint")) {

            LOGGER.log(Level.INFO, "unique constraint was violated: {0}", psqlException.getServerErrorMessage());
            response.status(Status.BAD_REQUEST_400);
            response.send("value already exists: " + psqlException.getServerErrorMessage());
            return null;
        }

        LOGGER.log(Level.WARNING, "caught error", realCause);
        response.status(Http.Status.INTERNAL_SERVER_ERROR_500);
        response.send("Failed to process request: " + realCause.getClass().getName() +
            "(" + realCause.getMessage() + ")");
        return null;
    }
}
