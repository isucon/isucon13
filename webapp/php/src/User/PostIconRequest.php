<?php

declare(strict_types=1);

namespace IsuPipe\User;

use JsonException;
use TypeError;
use UnexpectedValueException;

class PostIconRequest
{
    public function __construct(
        public string $image,
    ) {
    }

    /**
     * @throws UnexpectedValueException
     */
    public static function fromJson(string $json): PostIconRequest
    {
        try {
            $data = json_decode($json, flags: JSON_THROW_ON_ERROR);
        } catch (JsonException) {
            throw new UnexpectedValueException();
        }

        if (!isset($data->image)) {
            throw new UnexpectedValueException();
        }

        if (!is_string($data->image)) {
            throw new UnexpectedValueException();
        }

        $image = base64_decode($data->image);
        if ($image === false) {
            throw new UnexpectedValueException();
        }

        try {
            return new PostIconRequest(
                image: $image,
            );
        } catch (TypeError) {
            throw new UnexpectedValueException();
        }
    }
}
