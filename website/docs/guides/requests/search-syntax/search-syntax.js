db.runCommand({"aggregate": "collectionName", "pipeline": [{"$search": {"cosmosSearch": {"vector": "<vector>", "path": "<path>", "k": "<k>", "efSearch": "<efSearch>"}}}], "cursor": {}});
