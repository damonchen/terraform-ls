package lang

import (
	"fmt"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	tfjson "github.com/hashicorp/terraform-json"
)

// ParseBlock parses HCL configuration based on tfjson's SchemaBlock
// and keeps hold of all tfjson schema details on block or attribute level
func ParseBlock(block *hclsyntax.Block, labels []*Label, schema *tfjson.SchemaBlock) Block {
	b := &parsedBlock{
		hclBlock: block,
		labels:   labels,
	}
	if block == nil {
		return b
	}

	body := block.Body

	if schema == nil {
		b.unknownAttributes = body.Attributes
		b.unknownBlocks = body.Blocks
		return b
	}

	b.AttributesMap, b.unknownAttributes = parseAttributes(body.Attributes, schema.Attributes)
	b.BlockTypesMap, b.unknownBlocks = parseBlockTypes(body.Blocks, schema.NestedBlocks)

	return b
}

func AsHCLSyntaxBlock(block *hcl.Block) (*hclsyntax.Block, error) {
	if block == nil {
		return nil, nil
	}

	body, ok := block.Body.(*hclsyntax.Body)
	if !ok {
		return nil, fmt.Errorf("invalid configuration format: %T", block.Body)
	}

	bodyRng := body.Range()

	openBraceRng := hcl.Range{
		Filename: bodyRng.Filename,
		Start:    bodyRng.Start,
		// hclsyntax.Body range always starts with open brace
		End: hcl.Pos{
			Column: bodyRng.Start.Column + 1,
			Byte:   bodyRng.Start.Byte + 1,
			Line:   bodyRng.Start.Line,
		},
	}
	closeBraceRng := hcl.Range{
		Filename: bodyRng.Filename,
		// hclsyntax.Body range always ends with close brace
		Start: hcl.Pos{
			Column: bodyRng.End.Column - 1,
			Byte:   bodyRng.End.Byte - 1,
			Line:   bodyRng.End.Line,
		},
		End: bodyRng.End,
	}

	return &hclsyntax.Block{
		Type:        block.Type,
		TypeRange:   block.TypeRange,
		Labels:      block.Labels,
		LabelRanges: block.LabelRanges,

		OpenBraceRange:  openBraceRng,
		CloseBraceRange: closeBraceRng,

		Body: body,
	}, nil
}
