package ch.micha.deployment.controller.auth.logging;

import java.io.BufferedWriter;
import java.io.FileWriter;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.time.YearMonth;
import java.util.logging.Level;
import java.util.logging.Logger;

public class RequestLogFileWriter {
    private static final Logger LOGGER = Logger.getLogger(RequestLogFileWriter.class.getSimpleName());
    private static final String LOG_FILE_NAME_TEMPLATE = "access-log-%s.log";

    private BufferedWriter logOutput;
    private final Path logDir;

    private int currentMonth = YearMonth.now().getMonthValue();

    public RequestLogFileWriter(String logFilesDir) throws IOException {
        logDir = Paths.get(logFilesDir);
        initLogDir();
        validateLogFile();
    }

    public void writeLine(String line) throws IOException {
        LOGGER.log(Level.FINE, "writing new line: {0}", new Object[]{ line});

        if(YearMonth.now().getMonthValue() != currentMonth)
            validateLogFile();

        logOutput.write(line);
        logOutput.newLine();
        logOutput.flush();
    }

    private void initLogDir() throws IOException {
        if(!Files.exists(logDir)) {
            Files.createDirectories(logDir);
            LOGGER.log(Level.FINE, "created log file dir at {0}", new Object[]{ logDir.toAbsolutePath() });
        }
    }

    private void validateLogFile() throws IOException {
        currentMonth = YearMonth.now().getMonthValue();
        Path logFile = logDir.resolve(String.format(LOG_FILE_NAME_TEMPLATE, currentMonth));
        if(!Files.exists(logFile)) {
            LOGGER.log(Level.FINE, "log file invalid, creating new one at {0}", new Object[]{ logFile.toAbsolutePath() });
            if(logOutput != null)
                logOutput.close();

            Files.createFile(logFile);
            LOGGER.log(Level.FINE, "created new log file at {0}", new Object[]{ logFile.toAbsolutePath() });
        }
        logOutput = new BufferedWriter(new FileWriter(logFile.toFile(), true));
    }
}
