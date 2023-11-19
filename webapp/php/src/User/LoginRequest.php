<?php

declare(strict_types=1);

namespace IsuPipe\User;

use JsonException;
use TypeError;
use UnexpectedValueException;

class LoginRequest
{
    public function __construct(
        public string $username,
        // $password is non-hashed password.
        public string $password,
    ) {
    }

    /**
     * @throws UnexpectedValueException
     */
    public static function fromJson(string $json): LoginRequest
    {
        try {
            $data = json_decode($json, flags: JSON_THROW_ON_ERROR);
        } catch (JsonException $e) {
            throw new UnexpectedValueException(
                message: $e->getMessage(),
                previous: $e,
            );
        }

        if (
            !isset(
                $data->username,
                $data->password,
            )
        ) {
            throw new UnexpectedValueException('required fields are missing');
        }

        try {
            return new LoginRequest(
                username: $data->username,
                password: $data->password,
            );
        } catch (TypeError $e) {
            throw new UnexpectedValueException(
                message: $e->getMessage(),
                previous: $e,
            );
        }
    }
}
