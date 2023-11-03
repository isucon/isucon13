<?php

declare(strict_types=1);

namespace IsuPipe\Payment;

use IsuPipe\AbstractHandler;
use Psr\Http\Message\ResponseInterface as Response;
use Psr\Http\Message\ServerRequestInterface as Request;

class Handler extends AbstractHandler
{
    public function getPaymentResult(Request $request, Response $response): Response
    {
        // TODO: 実装
        return $response;
    }
}
