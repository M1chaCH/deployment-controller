package ch.micha.deployment.controller.auth.entity.page.addpage;

import io.helidon.common.Reflected;

@Reflected
public record AddPage(
    String url,
    String title,
    String description,
    boolean privateAccess
) { }
