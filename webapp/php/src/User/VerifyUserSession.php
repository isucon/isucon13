<?php

declare(strict_types=1);

namespace IsuPipe\User;

use Psr\Http\Message\ServerRequestInterface as Request;
use Slim\Exception\{ HttpException, HttpForbiddenException, HttpUnauthorizedException };
use SlimSession\Helper as Session;

trait VerifyUserSession
{
    const DEFAULT_SESSION_EXPIRES_KEY = 'EXPIRES';
    const DEFAULT_USER_ID_KEY = 'USERID';

    /**
     * @throws HttpException
     */
    protected function verifyUserSession(Request $request, Session $session): void
    {
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
