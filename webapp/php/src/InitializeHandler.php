<?php

namespace IsuPipe;

use Psr\Http\Message\ResponseInterface as Response;
use Psr\Http\Message\ServerRequestInterface as Request;

/**
 * FIXME: ポータルと足並み揃えて修正
 */
class InitializeHandler extends AbstractHandler
{
    public function __invoke(Request $request, Response $response): Response
    {
        // TODO: 実装
        return $response;
    }
}
