db.runCommand({
  createIndexes: '<collectionName>',
  indexes: [
    {
      name: '<indexName>',
      key: { '<path>': 'cosmosSearch' },
      cosmosSearchOptions: {
        kind: '<kind>',
        similarity: '<similarity>',
        dimensions: '<dimensions>',
        m: '<m>',
        efConstruction: '<efConstruction>',
        numLists: '<numLists>'
      }
    }
  ]
})
