package ch.micha.deployment.controller.auth.entity.user.adduser;

import io.helidon.common.Reflected;

@Reflected
public record AddUser(
    String mail,
    String password,
    boolean admin,
    boolean viewPrivate
) { }
