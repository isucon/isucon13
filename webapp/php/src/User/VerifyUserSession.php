<?php

declare(strict_types=1);

namespace IsuPipe\User;

use Psr\Http\Message\ServerRequestInterface as Request;
use Slim\Exception\{ HttpException, HttpForbiddenException, HttpInternalServerErrorException, HttpUnauthorizedException };
use SlimSession\Helper as Session;

trait VerifyUserSession
{
    protected const DEFAULT_SESSION_ID_KEY = 'SESSIONID';
    protected const DEFAULT_SESSION_EXPIRES_KEY = 'EXPIRES';
    protected const DEFAULT_USER_ID_KEY = 'USERID';
    protected const DEFAULT_USERNAME_KEY = 'USERNAME';

    /**
     * @throws HttpException
     */
    protected function verifyUserSession(Request $request, Session $session): void
    {
        if (
            session_set_cookie_params([
            'domain' => '*.u.isucon.dev',
            ]) === false
        ) {
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to set cookie params',
            );
        }
        if (session_name($this::DEFAULT_SESSION_ID_KEY) === false) {
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to set session name',
            );
        }
        if (session_start() === false) {
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to start session',
            );
        }

        $sessionExpires = $session->get($this::DEFAULT_SESSION_EXPIRES_KEY);
        if (!is_int($sessionExpires)) {
            throw new HttpForbiddenException(
                request: $request,
                message: 'failed to get EXPIRES value from session',
            );
        }

        $userId = $session->get($this::DEFAULT_USER_ID_KEY);
        if (!is_int($userId)) {
            throw new HttpUnauthorizedException(
                request: $request,
                message: 'failed to get USERID value from session',
            );
        }

        $now = time();
        if ($now > $sessionExpires) {
            throw new HttpUnauthorizedException(
                request: $request,
                message: 'session has expired',
            );
        }
    }
}
