<?php

namespace IsuPipe;

use JsonSerializable;

class InitializeResponse implements JsonSerializable
{
    public function __construct(
        public int $advertiseLevel,
        public string $language,
    ) {
    }

    /**
     * @return array<string, mixed>
     */
    public function jsonSerialize(): array
    {
        return [
            'advertise_level' => $this->advertiseLevel,
            'language' => $this->language,
        ];
    }
}
