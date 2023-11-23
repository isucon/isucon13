<?php

declare(strict_types=1);

namespace IsuPipe\Livestream;

class LivestreamViewerModel
{
    public function __construct(
        public ?int $userId = null,
        public ?int $livestreamId = null,
        public ?int $createdAt = null,
    ) {
    }
}
