package ch.micha.deployment.controller.auth.error;

public class InactiveUserException extends ForbiddenException {

    public InactiveUserException(String mail) {
        super("inactive user tried to access backend: %s".formatted(mail), "inactive user");
    }
}
