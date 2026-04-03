@ECHO OFF

SET freezetag-compose-path=%~p0..\compose.yaml

ECHO Killing the app.

WHERE /q docker compose
IF %ERRORLEVEL% NEQ 0 (
    docker-compose -f %freezetag-compose-path% down
) ELSE (
    docker compose -f %freezetag-compose-path% down
)

PAUSE
EXIT /B 0