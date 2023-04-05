package schema

// Auto generated GraphQL schema bundle
const schema = `
# Bytes32 is a 32 byte binary string, represented by 0x prefixed hexadecimal hash.
scalar Bytes32

# Address is a 20 byte Opera address, represented as 0x prefixed hexadecimal number.
scalar Address

# BigInt is a large integer value. Input is accepted as either a JSON number,
# or a hexadecimal string alternatively prefixed with 0x. Output is 0x prefixed hexadecimal.
scalar BigInt

# Long is a 64 bit unsigned integer value.
scalar Long

# Bytes is an arbitrary length binary string, represented as 0x-prefixed hexadecimal.
# An empty byte string is represented as '0x'.
scalar Bytes

# Cursor is a string representing position in a sequential list of edges.
scalar Cursor

# Time represents date and time including time zone information in RFC3339 format.
scalar Time

type Project {
    # Id of project
    id: Long!

    # List of contracts
    contracts: [ProjectContract!]!

    # Address of owner
    ownerAddress: Address!

    # Address to claim rewards
    receiverAddress: Address!

    # Name of project
    name: String!

    # URL of icon
    imageUrl: String!

    # URL
    url: String!

    # Sum of all transactions
    transactionsCount: Long!

    # Sum of claimed and pending rewards
    collectedRewards: Long!

    # Amount of tokens already received
    claimedRewards: Long!

    # Amount of tokens for claim
    rewardsToClaim: Long!
}
type ProjectContract {
    # Id of contract
    id: Long!

    # id of project
    projectId: Long!

    # address of contract
    address: Address!

    # approved contract
    approved: Boolean!
}
# Root schema definition
schema {
    query: Query
}

# Entry points for querying the API
type Query {
    # Value of all claimed tokens
    totalAmountClaimed: Long!

    # Value of all available tokens to claim
    totalAmountCollected: Long!

    # Version represents the API server version responding to your requests.
    totalTransactionCount: Long!

    # Returns the address of the gas monetization contract.
    gasMonetizationAddress: Address!

    # Projects represents list of validated projects
    projects: [Project!]!
}
`
