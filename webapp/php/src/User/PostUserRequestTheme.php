<?php

declare(strict_types=1);

namespace IsuPipe\User;

class PostUserRequestTheme
{
    public function __construct(
        public bool $darkMode,
    ) {
    }
}
