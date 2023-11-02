import * as fs from 'fs';
import * as path from 'path';
import { CodeGenerator } from '@himenon/openapi-typescript-code-generator';
import * as Templates from '@himenon/openapi-typescript-code-generator/templates';

const main = () => {
  const codeGenerator = new CodeGenerator(
    path.join(__dirname, '../../../docs/isupipe.yaml'),
  );

  const apiClientGeneratorTemplate = {
    generator: Templates.ClassApiClient.generator,
    option: {},
  };

  const typeDefCode = codeGenerator.generateTypeDefinition();
  const apiClientCode = codeGenerator.generateCode([
    {
      generator: () => [
        `import { Schemas, RequestBodies, Responses } from "./types";`,
      ],
    },
    codeGenerator.getAdditionalTypeDefinitionCustomCodeGenerator(),
    apiClientGeneratorTemplate,
  ]);

  fs.writeFileSync(__dirname + '/../../src/api/types.ts', typeDefCode, {
    encoding: 'utf-8',
  });
  fs.writeFileSync(__dirname + '/../../src/api/apiClient.ts', apiClientCode, {
    encoding: 'utf-8',
  });

  console.log('Generate API Client');
};

main();
