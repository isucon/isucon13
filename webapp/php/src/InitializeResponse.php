<?php

namespace IsuPipe;

use JsonSerializable;

class InitializeResponse implements JsonSerializable
{
    public function __construct(
        public string $language,
    ) {
    }

    /**
     * @return array<string, mixed>
     */
    public function jsonSerialize(): array
    {
        return [
            'language' => $this->language,
        ];
    }
}
