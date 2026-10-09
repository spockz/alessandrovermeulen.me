+++
# Start articles as drafts and derive suggested titles without publication-date prefixes.
date = '{{ .Date }}'
draft = true
title = {{ replace (replaceRE "^[0-9]{4}-[0-9]{2}-[0-9]{2}-" "" .File.ContentBaseName) "-" " " | title | jsonify }}
+++
