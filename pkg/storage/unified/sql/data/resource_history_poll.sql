SELECT
    {{ .Ident "resource_version" | .Into .Response.ResourceVersion }},
    {{ .Ident "namespace" | .Into .Response.Key.Namespace }},
    {{ .Ident "group" | .Into .Response.Key.Group }},
    {{ .Ident "resource" | .Into .Response.Key.Resource }},
    {{ .Ident "name" | .Into .Response.Key.Name }},
    {{ .Ident "folder" | .Into .Response.Folder }},
    {{ .Ident "value" | .Into .Response.Value }},
    {{ .Ident "action" | .Into .Response.Action }},
    {{ .Ident "previous_resource_version" | .Into .Response.PreviousRV }}

    FROM {{ .Ident "resource_history" }}
    WHERE 1 = 1
    AND {{ .Ident "group" }} = {{ .Arg .Group }}
    AND {{ .Ident "resource" }} = {{ .Arg .Resource }}
    AND {{ .Ident "resource_version" }} > {{ .Arg .SinceResourceVersion }}
    AND {{ .Ident "resource_version" }} <= (
        SELECT COALESCE(MIN(resource_version), {{ .CurrentEpoch }})
        FROM {{ .Ident "resource_lock" }}
        WHERE 1 = 1
        AND {{ .Ident "group" }} = {{ .Arg .Group }}
        AND {{ .Ident "resource" }} = {{ .Arg .Resource }}
    )
    ORDER BY {{ .Ident "resource_version" }} ASC
;
