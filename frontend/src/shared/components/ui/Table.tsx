/**
 * Copyright 2024 IBN Network (ICTU Blockchain Network)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { forwardRef } from 'react'
import type { HTMLAttributes } from 'react'
import { cn } from '@shared/utils/cn'

interface TableProps extends HTMLAttributes<HTMLTableElement> {
  striped?: boolean
}

const TableBase = forwardRef<HTMLTableElement, TableProps>(
  ({ className, striped = false, children, ...props }, ref) => {
    return (
      <div className="overflow-x-auto rounded-2xl border border-white/10 bg-black/30 backdrop-blur-xl shadow-inner shadow-black/40">
        <table
          ref={ref}
          className={cn(
            'w-full border-collapse text-white',
            striped && 'table-auto',
            className
          )}
          {...props}
        >
          {children}
        </table>
      </div>
    )
  }
)

TableBase.displayName = 'Table'

type TableHeaderProps = HTMLAttributes<HTMLTableSectionElement>

export const TableHeader = forwardRef<HTMLTableSectionElement, TableHeaderProps>(
  ({ className, children, ...props }, ref) => {
    return (
      <thead
        ref={ref}
        className={cn('bg-white/10 text-gray-200 uppercase tracking-widest text-xs', className)}
        {...props}
      >
        {children}
      </thead>
    )
  }
)

TableHeader.displayName = 'TableHeader'

interface TableBodyProps extends HTMLAttributes<HTMLTableSectionElement> {
  hover?: boolean
}

export const TableBody = forwardRef<HTMLTableSectionElement, TableBodyProps>(
  ({ className, hover = false, children, ...props }, ref) => {
    return (
      <tbody
        ref={ref}
        className={cn(
          'divide-y divide-white/10',
          hover && '[&>tr]:hover:bg-white/5',
          className
        )}
        {...props}
      >
        {children}
      </tbody>
    )
  }
)

TableBody.displayName = 'TableBody'

interface TableRowProps extends HTMLAttributes<HTMLTableRowElement> {
  striped?: boolean
}

export const TableRow = forwardRef<HTMLTableRowElement, TableRowProps>(
  ({ className, striped = false, children, ...props }, ref) => {
    return (
      <tr
        ref={ref}
        className={cn(
          'border-b border-white/10',
          striped && 'even:bg-white/5',
          className
        )}
        {...props}
      >
        {children}
      </tr>
    )
  }
)

TableRow.displayName = 'TableRow'

interface TableHeadProps extends HTMLAttributes<HTMLTableCellElement> {
  sortable?: boolean
}

export const TableHead = forwardRef<HTMLTableCellElement, TableHeadProps>(
  ({ className, sortable = false, children, ...props }, ref) => {
    return (
      <th
        ref={ref}
        className={cn(
          'px-6 py-3 text-left text-xs font-semibold text-gray-300 uppercase tracking-widest',
          sortable && 'cursor-pointer hover:bg-white/5',
          className
        )}
        {...props}
      >
        {children}
      </th>
    )
  }
)

TableHead.displayName = 'TableHead'

type TableCellProps = HTMLAttributes<HTMLTableCellElement>

export const TableCell = forwardRef<HTMLTableCellElement, TableCellProps>(
  ({ className, children, ...props }, ref) => {
    return (
      <td
        ref={ref}
        className={cn(
          'px-6 py-4 whitespace-nowrap text-sm text-gray-200',
          className
        )}
        {...props}
      >
        {children}
      </td>
    )
  }
)

TableCell.displayName = 'TableCell'

// Export compound component with proper typing
export const Table = Object.assign(TableBase, {
  Header: TableHeader,
  Body: TableBody,
  Row: TableRow,
  Head: TableHead,
  Cell: TableCell,
}) as typeof TableBase & {
  Header: typeof TableHeader
  Body: typeof TableBody
  Row: typeof TableRow
  Head: typeof TableHead
  Cell: typeof TableCell
}

