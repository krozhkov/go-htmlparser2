package dom

/** Types of elements found in htmlparser2's DOM */
type ElementType string

const (
	/** Type for the root element of a document */
	ElementTypeRoot ElementType = "root"
	/** Type for Text */
	ElementTypeText ElementType = "text"
	/** Type for <? ... ?> */
	ElementTypeDirective ElementType = "directive"
	/** Type for <!-- ... --> */
	ElementTypeComment ElementType = "comment"
	/** Type for <script> tags */
	ElementTypeScript ElementType = "script"
	/** Type for <style> tags */
	ElementTypeStyle ElementType = "style"
	/** Type for Any tag */
	ElementTypeTag ElementType = "tag"
	/** Type for <![CDATA[ ... ]]> */
	ElementTypeCDATA ElementType = "cdata"
	/** Type for <!doctype ...> */
	ElementTypeDoctype ElementType = "doctype"
)
