package ch.micha.deployment.controller.auth.resource;

import ch.micha.deployment.controller.auth.EncodingUtil;
import ch.micha.deployment.controller.auth.entity.user.adduser.AddUser;
import ch.micha.deployment.controller.auth.entity.user.edituser.EditUser;
import ch.micha.deployment.controller.auth.entity.user.User;
import ch.micha.deployment.controller.auth.error.AppRequestException;
import ch.micha.deployment.controller.auth.error.NotFoundException;
import ch.micha.deployment.controller.auth.logging.RequestLogHandler;
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
import java.util.logging.Level;
import java.util.logging.Logger;

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
        final String requestId = RequestLogHandler.parseRequestId(request);
        LOGGER.log(Level.FINE, "{0} loading all users", new Object[]{ requestId });

        response.send(db.execute(exec -> exec.namedQuery("select-users"))
                .map(item -> item.as(JsonObject.class)), JsonObject.class)
            .thenAccept(sentResponse -> LOGGER.log(Level.FINE, "{0} - successfully loaded all users", requestId))
            .exceptionally(t -> AppRequestException.respondFitting(response, requestId, t));
    }

    private void addUser(ServerRequest request, ServerResponse response, AddUser toAdd) {
        final String requestId = RequestLogHandler.parseRequestId(request);
        LOGGER.log(Level.FINE, "{0} adding user {1} as admin:{2}", new Object[]{ requestId, toAdd.mail(), toAdd.admin() });

        String salt = EncodingUtil.generateSalt();
        String hashedPassword = EncodingUtil.hashString(toAdd.password(), salt);
        final User user = new User(-1, toAdd.mail(), hashedPassword, salt, toAdd.admin(), toAdd.viewPrivate(),
            null, null);

        db.execute(exec -> exec
                .createNamedInsert("insert-user")
                .namedParam(user)
                .execute())
            .thenAccept(count -> {
                LOGGER.log(Level.FINE, "{0} added {1} user(s)", new Object[]{ requestId, count });
                response.status(Status.NO_CONTENT_204);
                response.send();
            })
            .exceptionally(t -> AppRequestException.respondFitting(response, requestId, t));
    }

    private void editUser(ServerRequest request, ServerResponse response, EditUser toEdit) {
        final String requestId = RequestLogHandler.parseRequestId(request);
        LOGGER.log(Level.FINE, "{0} updating user {1}-{2}", new Object[]{ requestId, toEdit.id(), toEdit.mail() });

        Optional<DbRow> userRow = db.execute(exec -> exec
            .createNamedGet("select-user")
            .addParam("id", toEdit.id())
            .execute()).await();

        if(userRow.isEmpty())
            throw new NotFoundException("could not find user with id " + toEdit.id(), "not found");

        User user = userRow.get().as(User.class);
        String newHashedPassword = EncodingUtil.hashString(toEdit.password(), user.salt());
        db.execute(exec -> exec
                .createNamedUpdate("update-user")
                .addParam("id", toEdit.id())
                .addParam("mail", toEdit.mail())
                .addParam("password", newHashedPassword)
                .addParam("admin", toEdit.admin())
                .addParam("view_private", toEdit.viewPrivate())
                .execute())
            .thenAccept(count -> {
                LOGGER.log(Level.FINE, "{0} changed {1} user(s)", new Object[]{ requestId, count });
                response.status(Status.NO_CONTENT_204);
                response.send();
            })
            .exceptionally(t -> AppRequestException.respondFitting(response, requestId, t));
    }

    private void deleteUser(ServerRequest request, ServerResponse response) {
        final String requestId = RequestLogHandler.parseRequestId(request);
        int userId = Integer.parseInt(request.path().param("id"));

        LOGGER.log(Level.FINE, "{0} deleting user with id {1}", new Object[]{ requestId, userId });

        db.execute(exec -> exec
                .createNamedDelete("delete-user")
                .addParam("id", userId)
                .execute())
            .thenAccept(count -> {
                LOGGER.log(Level.FINE, "{0} deleted {1} user(s)", new Object[]{ requestId, count });
                response.status(Status.NO_CONTENT_204);
                response.send();
            })
            .exceptionally(t -> AppRequestException.respondFitting(response, requestId, t));
    }
}
