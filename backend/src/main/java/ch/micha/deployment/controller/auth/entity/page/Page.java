package ch.micha.deployment.controller.auth.entity.page;

import io.helidon.common.Reflected;

@Reflected
public record Page(
    String id,
    String url,
    String title,
    String description,
    boolean privateAccess
) { }
