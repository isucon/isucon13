import argparse
import random

import markovify
import MeCab

def generate_positives(n=10):
    texts = []
    parsed_texts = []
    with open("./initial-data/positive_sentence.txt", "r") as f:
        for line in f.readlines():
            line = line.rstrip()
            if line:
                parsed = MeCab.Tagger('-Owakati').parse(line)
                texts.append(line)
                parsed_texts.append(parsed)
    if n <= len(texts):
        return texts[:n]

    remaining = n-len(texts)
    text_model = markovify.NewlineText('\n'.join(parsed_texts))
    for _ in range(remaining):
        sentence = text_model.make_short_sentence(140)
        if sentence is None:
            continue
        texts.append(sentence.replace(' ', ''))

    return texts


def generate_negatives(n=10):
    with open("./initial-data/negative_formats.txt", "r") as f:
        lines = f.readlines()
        formats = list(line.rstrip() for line in lines)

    with open('./initial-data/ngwords.txt', 'r') as f:
        ngwords = list(line.rstrip() for line in f.readlines())
        for _ in range(n):
            ngword = random.choice(ngwords)
            fmt = random.choice(formats)
            yield [fmt.format(word=ngword), ngword]

def main():
    def command_positive(args):
        for positive in generate_positives(args.n):
            print(positive)

    def command_negative(args):
        for negative in generate_negatives(args.n):
            print(negative)

    def command_help(args):
        print(parser.parse_args([args.command, '--help']))

    parser = argparse.ArgumentParser()
    subparsers = parser.add_subparsers()

    parser_positive = subparsers.add_parser('positive', help='ポジティブな文章生成')
    parser_positive.add_argument('-n', type=int, default=100, help='文章生成数')
    parser_positive.set_defaults(handler=command_positive)

    parser_negative = subparsers.add_parser('negative', help='ネガティブな文章生成')
    parser_negative.add_argument('-n', type=int, default=100, help='文章生成数')
    parser_negative.set_defaults(handler=command_negative)

    parser_help = subparsers.add_parser('help', help='see `help -h`')
    parser_help.add_argument('command', help='command name which help is shown')
    parser_help.set_defaults(handler=command_help)

    args = parser.parse_args()
    if hasattr(args, 'handler'):
        args.handler(args)
    else:
        parser.print_help()


if __name__ == '__main__':
    main()
