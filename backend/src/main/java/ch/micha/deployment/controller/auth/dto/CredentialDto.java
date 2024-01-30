package ch.micha.deployment.controller.auth.dto;

import io.helidon.common.Reflected;

@Reflected
public record CredentialDto(
    String mail,
    String password
) { }
