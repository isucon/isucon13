<?php

declare(strict_types=1);

namespace IsuPipe\Livecomment;

use JsonException;
use TypeError;
use UnexpectedValueException;

class ModerateRequest
{
    public function __construct(
        public string $ngWord,
    ) {
    }

    /**
     * @throws UnexpectedValueException
     */
    public static function fromJson(string $json): ModerateRequest
    {
        try {
            $data = json_decode($json, flags: JSON_THROW_ON_ERROR);
        } catch (JsonException $e) {
            throw new UnexpectedValueException(
                message: $e->getMessage(),
                previous: $e,
            );
        }

        if (!isset($data->ng_word)) {
            throw new UnexpectedValueException('required fields are missing');
        }

        try {
            return new ModerateRequest(
                ngWord: $data->ng_word,
            );
        } catch (TypeError $e) {
            throw new UnexpectedValueException(
                message: $e->getMessage(),
                previous: $e,
            );
        }
    }
}
