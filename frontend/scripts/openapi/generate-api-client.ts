import * as fs from 'fs';
import { CodeGenerator } from '@himenon/openapi-typescript-code-generator';
import * as Templates from '@himenon/openapi-typescript-code-generator/templates';

const main = () => {
  const codeGenerator = new CodeGenerator('scripts/openapi/spec.yaml');

  const apiClientGeneratorTemplate = {
    generator: Templates.ApiClient.generator,
    option: {},
  };

  const typeDefCode = codeGenerator.generateTypeDefinition();
  const apiClientCode = codeGenerator.generateCode([
    {
      generator: () => {
        return [`import { Schemas, RequestBodies } from "./types";`];
      },
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
