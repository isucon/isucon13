<?php

declare(strict_types=1);

namespace IsuPipe\User;

use JsonException;
use TypeError;
use UnexpectedValueException;

class PostUserRequest
{
    public function __construct(
        public string $name,
        public string $displayName,
        public string $description,
        // $password is non-hashed password.
        public string $password,
        public PostUserRequestTheme $theme,
    ) {
    }

    /**
     * @throws UnexpectedValueException
     */
    public static function fromJson(string $json): PostUserRequest
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
                $data->name,
                $data->display_name,
                $data->description,
                $data->password,
                $data->theme,
                $data->theme->dark_mode,
            )
        ) {
            throw new UnexpectedValueException('required fields are missing');
        }

        try {
            return new PostUserRequest(
                name: $data->name,
                displayName: $data->display_name,
                description: $data->description,
                password: $data->password,
                theme: new PostUserRequestTheme(
                    darkMode:  $data->theme->dark_mode,
                ),
            );
        } catch (TypeError $e) {
            throw new UnexpectedValueException(
                message: $e->getMessage(),
                previous: $e,
            );
        }
    }
}
