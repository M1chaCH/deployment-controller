package ch.micha.deployment.controller.auth.entity.user;

import io.helidon.common.Reflected;
import java.time.LocalDateTime;

@Reflected
public record User (
    int id,
    String mail,
    String password,
    String salt,
    boolean admin,
    boolean viewPrivate,
    LocalDateTime createdAt,
    LocalDateTime lastLoginAt
) { }
