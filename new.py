#!/usr/bin/env python3


from argparse import ArgumentParser
from pathlib import Path
from subprocess import run


def controller(name: str):
    type_name = name.title()
    code = (
        f"""
	package {name}

import (
	"context"
	"net/http"

	"github.com/reddec/dreaming-bard/internal/dreamwriter"
)


func New(dreamer *dreamwriter.DreamWriter) *{type_name} {{
	mux := http.NewServeMux()
	srv := &{type_name}{{
		Handler:  mux,
		dreamer:  dreamer,
	}}
	mux.HandleFunc("GET /", srv.index)

	return srv
}}

type {type_name} struct {{
	http.Handler
    dreamer *dreamwriter.DreamWriter
}}

func (srv *{type_name}) Run(ctx context.Context) error {{
	// TODO: background tasks
	return nil
}}

func (srv *{type_name}) index(w http.ResponseWriter, r *http.Request) {{
	viewIndex().HTML(w, indexParams{{}})
}}
""".strip()
        + "\n"
    )
    root = Path(__file__).parent / "internal" / "server" / name
    print(root)
    root.mkdir(exist_ok=True, parents=True)
    views = root / "views"
    views.mkdir(exist_ok=True, parents=True)
    index_view = views / "index.gohtml"
    if not index_view.exists():
        index_view.write_text(
            """
{{define "main"}}
    <h1>"""
            + name
            + """</h1>
{{end}}
""".strip()
            + "\n"
        )

    controller_file = root / (name + ".go")
    if not controller_file.exists():
        controller_file.write_text(code)

    run(["go", "generate", "./..."])



def main():
    parser = ArgumentParser()
    sub = parser.add_subparsers(required=True)
    
    ctrl = sub.add_parser('controller',help='Create new controller')
    ctrl.add_argument('name')
    ctrl.set_defaults(func=lambda args: controller(args.name))

    args = parser.parse_args()
    args.func(args)


if __name__ == '__main__':
    main()
