package ch.micha.deployment.controller.auth.dto;

import io.helidon.common.Reflected;
import java.util.UUID;

@Reflected
public record EditUserDto(
    UUID id,
    String mail,
    String password,
    boolean admin,
    String[] pagesToAllow,
    String[] pagesToDisallow
) { }
