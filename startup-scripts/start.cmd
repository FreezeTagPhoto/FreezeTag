@ECHO OFF

SET freezetag-compose-path=%~p0..\compose.yaml

WHERE /q docker
IF %ERRORLEVEL% NEQ 0 (
  ECHO docker could not be found. Please install docker: https://docs.docker.com/engine/install/ >&2
  PAUSE
  EXIT /B 1
)
ECHO Found docker.

WHERE /q docker compose
IF %ERRORLEVEL% NEQ 0 (
  ECHO docker compose could not be found. Checking for docker-compose [legacy] >&2
  WHERE /q docker-compose
  IF %ERRORLEVEL% NEQ 0 (
    ECHO docker compose could not be found. Please install docker compose: https://docs.docker.com/compose/install/ >&2
    ECHO you already have docker, docker compose is a plugin on top of docker >&2
    PAUSE
    EXIT /B 1
  )
  ECHO Found docker-compose.
  ECHO warning: docker-compose is a legacy version of docker compose, and may not work correctly >&2
  ECHO Consider updating: https://docs.docker.com/compose/install/ >&2

  ECHO Running the app. If this is the first time, this may take a moment to build all requirements.
  ECHO Run build.bat if you need to import new features after pulling a new version of the repo.
  docker-compose -f %freezetag-compose-path% up -d
) ELSE (
    ECHO Found docker compose.

    ECHO Running the app. If this is the first time, this may take a moment to build all requirements.
    ECHO Run build.bat if you need to import new features after pulling a new version of the repo.
    docker compose -f %freezetag-compose-path% up -d
)

PAUSE
EXIT /B 0