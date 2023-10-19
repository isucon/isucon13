<?php

declare(strict_types=1);

namespace IsuPipe\Payment;

use Psr\Http\Message\ResponseInterface as Response;
use Psr\Http\Message\ServerRequestInterface as Request;

class Handler
{
    public function getPaymentResult(Request $request, Response $response): Response
    {
        // TODO: 実装
        return $response;
    }
}
