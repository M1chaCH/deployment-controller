package ch.micha.deployment.controller.auth.dto;

import io.helidon.common.Reflected;

@Reflected
public record ChangeCredentialDto(
    String mail,
    String oldPassword,
    String password
) { }
