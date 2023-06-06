package ch.micha.deployment.controller.auth.entity.credentials;

import io.helidon.common.Reflected;

@Reflected
public record Credential(
    String mail,
    String password
) { }
