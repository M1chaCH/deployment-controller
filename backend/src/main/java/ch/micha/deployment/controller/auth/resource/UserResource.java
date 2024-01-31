package ch.micha.deployment.controller.auth.resource;

import ch.micha.deployment.controller.auth.EncodingUtil;
import ch.micha.deployment.controller.auth.db.CachedUserDb;
import ch.micha.deployment.controller.auth.db.UserEntity;
import ch.micha.deployment.controller.auth.dto.EditUserDto;
import ch.micha.deployment.controller.auth.dto.UserReadDto;
import ch.micha.deployment.controller.auth.error.AppRequestException;
import ch.micha.deployment.controller.auth.error.BadRequestException;
import ch.micha.deployment.controller.auth.error.NotFoundException;
import ch.micha.deployment.controller.auth.logging.RequestLogHandler;
import io.helidon.common.http.Http.Status;
import io.helidon.webserver.Handler;
import io.helidon.webserver.Routing.Rules;
import io.helidon.webserver.ServerRequest;
import io.helidon.webserver.ServerResponse;
import io.helidon.webserver.Service;
import java.sql.SQLException;
import java.util.List;
import java.util.UUID;
import java.util.logging.Level;
import java.util.logging.Logger;

public class UserResource implements Service{
    private static final Logger LOGGER = Logger.getLogger(UserResource.class.getSimpleName());

    private final CachedUserDb db;

    public UserResource(CachedUserDb db) {
        this.db = db;
    }

    @Override
    public void update(Rules rules) {
        rules
            .get("/", this::getUsers)
            .post("/", Handler.create(EditUserDto.class, this::addUser))
            .put("/", Handler.create(EditUserDto.class, this::editUser))
            .delete("/{id}", this::deleteUser);
    }

    private void getUsers(ServerRequest request, ServerResponse response) {
        final String requestId = RequestLogHandler.parseRequestId(request);
        LOGGER.log(Level.FINE, "{0} loading all users", new Object[]{ requestId });

        try {
            List<UserReadDto> users = db.selectUsers().stream().map(UserEntity::asDto).toList();

            response.send(users)
                    .thenAccept(sentResponse -> LOGGER.log(Level.FINE, "{0} - successfully loaded all users", requestId))
                    .exceptionally(t -> AppRequestException.respondFitting(response, requestId, t));
        } catch (SQLException e) {
            throw new BadRequestException("unhandled SQL exception", "could not load users, db error: " + e.getMessage(), e);
        }
    }

    private void addUser(ServerRequest request, ServerResponse response, EditUserDto toAdd) {
        final String requestId = RequestLogHandler.parseRequestId(request);
        LOGGER.log(Level.FINE, "{0} adding user {1} as admin:{2}", new Object[]{ requestId, toAdd.mail(), toAdd.admin() });

        String salt = EncodingUtil.generateSalt();
        String hashedPassword = EncodingUtil.hashString(toAdd.password(), salt);

        try {
            db.insertUser(toAdd.id(), toAdd.mail(), hashedPassword, salt, toAdd.admin(), toAdd.pagesToAllow());

            LOGGER.log(Level.FINE, "{0} added user", new Object[]{ requestId });
            response.status(Status.NO_CONTENT_204);
            response.send();
        } catch (SQLException e) {
            throw new BadRequestException("unhandled sql exception", "could not add user, db error: " + e.getMessage(), e);
        }
    }

    private void editUser(ServerRequest request, ServerResponse response, EditUserDto toEdit) {
        final String requestId = RequestLogHandler.parseRequestId(request);
        LOGGER.log(Level.FINE, "{0} updating user {1}", new Object[]{ requestId, toEdit.id() });

        UserEntity existingUser;
        try {
            existingUser = db.selectUser(toEdit.id());
        } catch (SQLException e) {
            throw new NotFoundException("user by id %s was not found".formatted(toEdit.id()), "could not find user");
        }

        if(existingUser.isAdmin() && !toEdit.admin() && db.countAdminUsers() <= 1) {
            throw new BadRequestException("user tried to remove admin rights from the last admin", "one admin required");
        }

        try {
            String hashedNewPassword = toEdit.password().isBlank() ? existingUser.getPassword() : EncodingUtil.hashString(toEdit.password(), existingUser.getSalt());

            db.updateUserWithPages(toEdit.id(), hashedNewPassword, toEdit.admin(), toEdit.pagesToAllow(), toEdit.pagesToDisallow());

            LOGGER.log(Level.FINE, "{0} edited user", new Object[]{ requestId });
            response.status(Status.NO_CONTENT_204);
            response.send();
        } catch (SQLException e) {
            throw new BadRequestException("unexpected SQL error", "could not edit user, db error: " + e.getMessage(), e);
        }
    }

    private void deleteUser(ServerRequest request, ServerResponse response) {
        final String requestId = RequestLogHandler.parseRequestId(request);
        UUID userId = UUID.fromString(request.path().param("id"));

        LOGGER.log(Level.FINE, "{0} deleting user with id {1}", new Object[]{ requestId, userId });

        UserEntity existingUser;
        try {
            existingUser = db.selectUser(userId);
        } catch (SQLException e) {
            return;
        }

        if(existingUser.isAdmin() && db.countAdminUsers() <= 1) {
            throw new BadRequestException("user tried to delete last admin", "one admin user is required");
        }

        try {
            db.deleteUser(userId);
            LOGGER.log(Level.FINE, "{0} deleted user", new Object[]{ requestId });
            response.status(Status.NO_CONTENT_204);
            response.send();
        } catch (SQLException e) {
            throw new BadRequestException("unexpected db error", "user could not be deleted, db error: " + e.getMessage(), e);
        }
    }
}
