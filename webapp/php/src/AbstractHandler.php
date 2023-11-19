<?php

declare(strict_types=1);

namespace IsuPipe;

use JsonSerializable;
use Psr\Http\Message\ResponseInterface as Response;
use RuntimeException;

abstract class AbstractHandler
{
    protected function jsonResponse(Response $response, array|JsonSerializable $data, int $status = 200): Response
    {
        $response->getBody()->write(json_encode($data));

        return $response->withHeader('Content-Type', 'application/json;charset=utf-8')
            ->withStatus($status);
    }

    /**
     * @param array<string, mixed> $array
     */
    protected function getAsInt(array $array, string $key): int|false
    {
        if (!isset($array[$key])) {
            return false;
        }

        $value = filter_var($array[$key], FILTER_VALIDATE_INT);
        if (!is_int($value)) {
            return false;
        }

        return $value;
    }

    /**
     * @param list<string> $command
     * @throws RuntimeException
     */
    protected function execCommand(array $command): void
    {
        $fp = fopen('php://temp', 'w+');
        $descriptorSpec = [
            1 => $fp,
            2 => $fp,
        ];

        $process = proc_open($command, $descriptorSpec, $_);
        if ($process === false) {
            throw new RuntimeException('cannot open process');
        }
        if (proc_close($process) !== 0) {
            rewind($fp);
            throw new RuntimeException(stream_get_contents($fp) ?: '');
        }
    }
}
