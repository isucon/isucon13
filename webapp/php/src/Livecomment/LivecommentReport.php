<?php

declare(strict_types=1);

namespace IsuPipe\Livecomment;

use JsonSerializable;
use IsuPipe\User\User;

class LivecommentReport implements JsonSerializable
{
    public function __construct(
        public int $id,
        public User $reporter,
        public Livecomment $livecomment,
        public int $createdAt,
    ) {
    }

    /**
     * @return array<string, mixed>
     */
    public function jsonSerialize(): array
    {
        return [
            'id' => $this->id,
            'reporter' => $this->reporter,
            'livecomment' => $this->livecomment,
            'created_at' => $this->createdAt,
        ];
    }
}
