package ch.micha.deployment.controller.auth.entity.user.edituser;

import io.helidon.common.Reflected;

@Reflected
public record EditUser(
    int id,
    String mail,
    String password,
    boolean admin,
    boolean viewPrivate
) { }
