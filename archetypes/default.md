+++
# New articles start as drafts so a preview cannot accidentally publish unfinished writing.
date = '{{ .Date }}'
draft = true
title = '{{ replace .File.ContentBaseName "-" " " | title }}'
+++
