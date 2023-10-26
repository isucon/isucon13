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
        } catch (JsonException) {
            throw new UnexpectedValueException();
        }

        if (!isset($data->ng_word)) {
            throw new UnexpectedValueException();
        }

        try {
            return new ModerateRequest(
                ngWord: $data->ng_word,
            );
        } catch (TypeError) {
            throw new UnexpectedValueException();
        }
    }
}
