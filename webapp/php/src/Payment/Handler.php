<?php

declare(strict_types=1);

namespace IsuPipe\Payment;

use IsuPipe\AbstractHandler;
use PDO;
use PDOException;
use Psr\Http\Message\ResponseInterface as Response;
use Psr\Http\Message\ServerRequestInterface as Request;
use Slim\Exception\HttpInternalServerErrorException;

class Handler extends AbstractHandler
{
    public function __construct(
        private PDO $db,
    ) {
    }

    public function getPaymentResult(Request $request, Response $response): Response
    {
        $this->db->beginTransaction();

        try {
            $stmt = $this->db->prepare('SELECT IFNULL(SUM(tip), 0) FROM livecomments');
            $stmt->execute();
            $totalTip = (int) $stmt->fetchColumn();
        } catch (PDOException $e) {
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to count total tip: ' . $e->getMessage(),
                previous: $e,
            );
        }

        $this->db->commit();

        return $this->jsonResponse($response, new PaymentResult(
            totalTip: $totalTip
        ));
    }
}
