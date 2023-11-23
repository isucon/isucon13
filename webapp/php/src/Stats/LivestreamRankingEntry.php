<?php

declare(strict_types=1);

namespace IsuPipe\Stats;

/**
 * @phpstan-type LivestreamRanking list<LivestreamRankingEntry>
 */
class LivestreamRankingEntry
{
    public function __construct(
        public int $livestreamId,
        public int $score,
    ) {
    }

    public function compare(LivestreamRankingEntry $other): int
    {
        if ($this->score === $other->score) {
            return $this->livestreamId <=> $other->livestreamId;
        }

        return $this->score <=> $other->score;
    }
}
