package ch.micha.deployment.controller.auth;

import ch.micha.deployment.controller.auth.auth.AuthHandler;
import ch.micha.deployment.controller.auth.auth.AuthService;
import ch.micha.deployment.controller.auth.db.CachedUserDb;
import ch.micha.deployment.controller.auth.db.CachedPageDb;
import ch.micha.deployment.controller.auth.db.UserPageDb;
import ch.micha.deployment.controller.auth.error.AppRequestException;
import ch.micha.deployment.controller.auth.error.GlobalErrorHandler;
import ch.micha.deployment.controller.auth.logging.RequestLogHandler;
import ch.micha.deployment.controller.auth.mail.SendMailDto;
import ch.micha.deployment.controller.auth.mail.SendMailProcessor;
import ch.micha.deployment.controller.auth.resource.ContactResource;
import ch.micha.deployment.controller.auth.resource.PageResource;
import ch.micha.deployment.controller.auth.resource.UserResource;
import io.helidon.common.LogConfig;
import io.helidon.common.reactive.Single;
import io.helidon.config.Config;
import io.helidon.media.jsonb.JsonbSupport;
import io.helidon.media.jsonp.JsonpSupport;
import io.helidon.webserver.HttpException;
import io.helidon.webserver.Routing;
import io.helidon.webserver.WebServer;
import io.helidon.webserver.cors.CorsSupport;
import io.helidon.webserver.cors.CrossOriginConfig;
import java.sql.Connection;
import java.sql.DriverManager;
import java.sql.SQLException;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.LinkedBlockingQueue;
import java.util.logging.Level;
import java.util.logging.Logger;

public final class Main {
    private static final Logger LOGGER = Logger.getLogger(Main.class.getSimpleName());

    private Main() {
    }

    public static void main(final String[] args) {
        startServer();
    }

    /**
     * Start the server.
     *
     * @return the created {@link WebServer} instance
     */
    static Single<WebServer> startServer() {
        // load logging configuration
        LogConfig.configureRuntime();

        // By default, this will pick up application.yaml from the classpath
        Config config = Config.create();

        WebServer server = WebServer.builder(createRouting(config))
            .config(config.get("server"))
            .addMediaSupport(JsonpSupport.create())
            .addMediaSupport(JsonbSupport.create())
            .build();

        Single<WebServer> webserver = server.start();

        webserver.forSingle(ws -> {
            LOGGER.log(Level.INFO, "server up and running at http://localhost:{0}", ws.port());
            ws.whenShutdown().thenRun(() -> LOGGER.info("server shut down! enjoy your resources (:"));
        }).exceptionallyAccept(t -> LOGGER.log(Level.SEVERE, "failed to start server:", t));

        return webserver;
    }

    /**
     * Creates new {@link Routing}.
     *
     * @param config configuration of this server
     * @return routing configured with JSON support, a health check, and a service
     */
    private static Routing createRouting(Config config) {
        Config appConfig = config.get("app");

        String clientUrl = appConfig.get("security.frontend").asString().get();
        final CorsSupport corsSupport = CorsSupport.builder()
            .addCrossOrigin(CrossOriginConfig.builder()
                .pathPattern("*")
                .allowOrigins(clientUrl)
                .allowCredentials(true)
                .allowMethods("GET", "PUT", "POST", "DELETE", "OPTIONS")
                .build())
            .build();

        String dbUrl = config.get("db.connection.url").asString().get();
        String dbUser = config.get("db.connection.username").asString().get();
        String dbPassword = config.get("db.connection.password").asString().get();
        Connection dbConnection = null;
        try {
            dbConnection = DriverManager.getConnection(dbUrl, dbUser, dbPassword);
        } catch (SQLException e) {
            LOGGER.log(Level.SEVERE, "could not connect to Db, stopping -- {0} : {1}", new Object[]{ dbUrl, dbUser });
            System.exit(99);
        }

        UserPageDb userPageDb = new UserPageDb(dbConnection);
        CachedUserDb userDb = new CachedUserDb(dbConnection, userPageDb);
        CachedPageDb pageDb = new CachedPageDb(dbConnection, userDb);

        BlockingQueue<SendMailDto> sendMailQueue = prepareMailQueue(appConfig);
        GlobalErrorHandler errorHandler = new GlobalErrorHandler();
        RequestLogHandler requestLogHandler = new RequestLogHandler(appConfig);
        AuthService authService = new AuthService(userDb, pageDb, appConfig, sendMailQueue);
        AuthHandler authHandler = new AuthHandler(authService);

        Routing.Builder builder = Routing.builder()
            .error(AppRequestException.class, errorHandler::handleAppRequestException)
            .error(HttpException.class, errorHandler::handleHttpException)
            .error(Exception.class, errorHandler::handleException)
            .any(requestLogHandler)
            .any(corsSupport)
            .register("/security", authService)
            .any("/users", authHandler)
            .any("/pages", authHandler)
            .register("/users", new UserResource(userDb, sendMailQueue, appConfig.get("security")))
            .register("/pages", new PageResource(pageDb, userDb))
            .register("/contact", new ContactResource(sendMailQueue, appConfig.get("security").get("default").get("admin").asString().get()));

        return builder.build();
    }

    private static BlockingQueue<SendMailDto> prepareMailQueue(Config appConfig) {
        BlockingQueue<SendMailDto> sendMailQueue = new LinkedBlockingQueue<>();
        Thread sendMailThread = new Thread(new SendMailProcessor(sendMailQueue, appConfig));
        sendMailThread.setName("mail-sender");
        sendMailThread.start();
        return sendMailQueue;
    }
}
