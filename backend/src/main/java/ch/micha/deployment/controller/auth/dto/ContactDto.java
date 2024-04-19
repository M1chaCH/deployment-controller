package ch.micha.deployment.controller.auth.dto;

import io.helidon.common.Reflected;

@Reflected
public record ContactDto(
    String mail,
    String message
) { }
